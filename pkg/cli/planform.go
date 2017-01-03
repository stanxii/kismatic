package cli

import (
	"fmt"
	"io"
	"strconv"

	"strings"

	gc "github.com/apprenda/goncurses"
	"github.com/apprenda/kismatic/pkg/install"
)

var (
	helpMap        = make(map[*gc.Field]string)
	countFields    = make(map[*gc.Field]bool)
	listFields     = make(map[*gc.Field]int)
	storageOptions = []string{"NONE", "NFS", "CLUSTER"}
)

func PaintPlanForm(out io.Writer, planner install.Planner, planFile string) {
	stdscr, _ := gc.Init()
	defer gc.End()

	maxy, maxx := stdscr.MaxYX()
	if maxx < 44 || maxy < 12 {
		fmt.Printf("Screen size too small for interactive mode. (%vx%v); your best bet is at least 80x40 \n", maxx, maxy)
		return
	}
	fields := make([]*gc.Field, 8)

	gc.Echo(false)
	gc.CBreak(true)
	gc.StartColor()
	stdscr.Keypad(true)

	gc.InitPair(1, gc.C_WHITE, gc.C_BLACK)
	gc.InitPair(2, gc.C_BLACK, gc.C_WHITE)
	gc.InitPair(3, gc.C_BLUE, gc.C_WHITE)
	gc.InitPair(4, gc.C_WHITE, gc.C_BLUE)

	fields[0], _ = makeNumberField(2, "3")
	fields[1], _ = makeNumberField(4, "2")
	fields[2], _ = makeNumberField(6, "3")
	fields[3], _ = makeNumberField(8, "2")
	fields[4], _ = makeListField(10, "0")
	fields[5], _ = makeNumberField(12, "0")
	fields[5].SetOptionsOff(gc.FO_EDIT | gc.FO_VISIBLE)
	fields[6], _ = makeButton(14, "Generate")
	goBox := fields[6]
	fields[7], _ = makeInfoBox(12, int32(maxx), 38)
	helpBox := fields[7]

	form, _ := gc.NewForm(fields)
	form.Post()
	defer form.UnPost()
	defer form.Free()
	for _, f := range fields {
		defer f.Free()
	}

	stdscr.AttrOn(gc.ColorPair(1) | gc.A_BOLD)
	stdscr.MovePrint(0, 2, "Generate a Plan for your Kubernetes Cluster")

	stdscr.AttrOff(gc.A_BOLD)
	stdscr.MovePrint(2, 2, "Number of Etcd nodes:")
	helpMap[fields[0]] = `Etcd nodes are used to store data Kubernetes needs to find and monitor workloads. More nodes makes this data safer.

Count  Safe for
1      Unsafe. Use only for small development clusters
3      Failure of any one node
5      Simultaneous failure of two nodes
7      Simultaneous failure of three nodes`

	stdscr.MovePrint(4, 2, "Number of Master nodes:")
	helpMap[fields[1]] = `Master nodes monitor and control workloads on the Kubernetes cluster.

Count  Safe for
1      Unsafe. Use only for small development clusters
2+     Failure of any one node`

	stdscr.MovePrint(6, 2, "Number of Worker nodes:")
	helpMap[fields[2]] = `Worker nodes are where most workloads are run in Kubernetes.

Count  Safe for
1      Unsafe. Use only for small development clusters
2+     Failure of worker nodes`

	stdscr.MovePrint(8, 2, "Number of Ingress nodes:")
	helpMap[fields[3]] = `Ingress nodes are a special kind of worker that open up HTTP access to workloads inside the cluster to clients that aren't part of the cluster. 

Count  Safe for
0      You don't want Kismatic managed Ingress
1      Unsafe. Use only for small development clusters
2+     Failure of any one node`

	stdscr.MovePrint(10, 2, "Persistent storage:")
	helpMap[fields[4]] = `To use persistent features of Kismatic, such as long term health monitoring or log aggregation, you will need a storage solution.

Value    Meaning
NONE     You don't intend to use any persistent features
NFS      You have an NFS file server and will provide one or more mountable shares
CLUSTER  You want Kismatic to build you a file storage cluster`

	stdscr.MovePrint(12, 2, "Number of storage nodes:")
	helpMap[fields[5]] = `If Kismatic is going to build a storage cluster for you, you will need to provide nodes for the data to be hosted from.

Count  Safe for
0      No store or to add your own NFS Shares
1      Test or demo cluster
2      Replicated storage or distributed storage without replicas
4+     Distributed storage with distributed replicas`
	helpMap[fields[6]] = `Ready to rock!
	
Press ENTER and a plan will be generated with placeholders for the nodes you specified above.`
	stdscr.Refresh()

	form.Driver(gc.REQ_FIRST_FIELD)
	form.Driver(gc.REQ_END_LINE)
	showHelp(&form, helpBox, helpMap)

	ch := stdscr.GetChar()

editloop:
	for ch != 'q' && ch != 0x1b {
		switch ch {
		case gc.KEY_UP:
			field, _ := form.Current()
			if _, ok := countFields[field]; ok {
				cur := strings.TrimSpace(field.Buffer())
				intval, err := strconv.Atoi(cur)
				if err == nil && intval >= 0 {
					intval = intval + 1
				} else {
					intval = 0
				}
				field.SetBuffer(strconv.Itoa(intval) + " ")
			} else if cur, ok := listFields[field]; ok {
				if cur > 0 {
					cur = cur - 1
				} else {
					cur = 0
				}

				listFields[field] = cur
				if cur == 2 {
					fields[5].SetOptionsOn(gc.FO_EDIT | gc.FO_VISIBLE)
				} else {
					fields[5].SetOptionsOff(gc.FO_EDIT | gc.FO_VISIBLE)
				}
				field.SetBuffer(storageOptions[cur])
			}
		case gc.KEY_DOWN:
			field, _ := form.Current()
			if _, ok := countFields[field]; ok {
				field, _ := form.Current()
				cur := strings.TrimSpace(field.Buffer())
				intval, err := strconv.Atoi(cur)
				if err == nil && intval > 0 {
					intval = intval - 1
				} else {
					intval = 0
				}
				field.SetBuffer(strconv.Itoa(intval) + " ")
			} else if cur, ok := listFields[field]; ok {
				if cur < len(storageOptions)-1 {
					cur = cur + 1
				} else {
					cur = len(storageOptions) - 1
				}

				listFields[field] = cur
				if cur == 2 {
					fields[5].SetOptionsOn(gc.FO_EDIT | gc.FO_VISIBLE)
				} else {
					fields[5].SetOptionsOff(gc.FO_EDIT | gc.FO_VISIBLE)
				}
				field.SetBuffer(storageOptions[cur])
			}
		case gc.KEY_TAB:
			form.Driver(gc.REQ_NEXT_FIELD)
			form.Driver(gc.REQ_END_LINE)

			showHelp(&form, helpBox, helpMap)
		case 0x161: //shift-tab
			form.Driver(gc.REQ_PREV_FIELD)
			form.Driver(gc.REQ_END_LINE)
			showHelp(&form, helpBox, helpMap)
		case gc.KEY_ENTER, gc.KEY_RETURN:
			field, _ := form.Current()

			if field == goBox {
				break editloop
			}

			form.Driver(gc.REQ_NEXT_FIELD)
			form.Driver(gc.REQ_END_LINE)
			showHelp(&form, helpBox, helpMap)
		}
		stdscr.Refresh()
		ch = stdscr.GetChar()
	}

	gc.End()

	etcdNodes, _ := strconv.Atoi(strings.TrimSpace(fields[0].Buffer()))
	masterNodes, _ := strconv.Atoi(strings.TrimSpace(fields[1].Buffer()))
	workerNodes, _ := strconv.Atoi(strings.TrimSpace(fields[2].Buffer()))
	ingressNodes, _ := strconv.Atoi(strings.TrimSpace(fields[3].Buffer()))

	fmt.Fprintf(out, "Generating installation plan file with %d etcd nodes, %d master nodes, %d worker nodes and %d ingress nodes\n",
		etcdNodes, masterNodes, workerNodes, ingressNodes)

	plan := buildPlan(etcdNodes, masterNodes, workerNodes, ingressNodes)

	install.WritePlanTemplate(plan, planner)

	fmt.Fprintf(out, "Edit the file to further describe your cluster. Once ready, execute the \"install validate \" command to proceed\n")

}

func showHelp(form *gc.Form, helpBox *gc.Field, helpMap map[*gc.Field]string) {
	field, _ := form.Current()
	if help, ok := helpMap[field]; ok {
		helpBox.SetBuffer(help)
	} else {
		helpBox.SetBuffer("")
	}
}

func makeNumberField(top int32, def string) (*gc.Field, error) {
	field, oth := gc.NewField(1, 4, top, 29, 0, 0)
	field.SetForeground(gc.ColorPair(2))
	field.SetBackground(gc.ColorPair(2) | gc.A_UNDERLINE | gc.A_BOLD)
	field.SetOptionsOff(gc.FO_AUTOSKIP)
	// field.SetJustification(3)
	field.SetBuffer(def)

	countFields[field] = true

	return field, oth
}

func makeListField(top int32, def string) (*gc.Field, error) {
	field, oth := gc.NewField(1, 7, top, 29, 0, 0)
	field.SetForeground(gc.ColorPair(2))
	field.SetBackground(gc.ColorPair(2) | gc.A_UNDERLINE | gc.A_BOLD)
	field.SetOptionsOff(gc.FO_AUTOSKIP)
	// field.SetJustification(3)
	field.SetBuffer(def)

	listFields[field] = 0

	field.SetBuffer(storageOptions[0])

	return field, oth
}

func makeButton(top int32, text string) (*gc.Field, error) {
	field, oth := gc.NewField(1, int32(len(text))+1, top, 25, 0, 0)
	field.SetBackground(gc.ColorPair(3) | gc.A_BOLD | gc.A_STANDOUT)
	field.SetOptionsOff(gc.FO_AUTOSKIP)
	field.SetOptionsOff(gc.FO_EDIT)
	field.SetBuffer(text)

	return field, oth
}

func makeInfoBox(height, maxx, left int32) (*gc.Field, error) {
	top := int32(2)
	width := int32(maxx - 2 - left)
	field, oth := gc.NewField(height-top, width, top, left, 0, 0)
	field.SetBackground(gc.ColorPair(4) | gc.A_NORMAL)
	field.SetOptionsOff(gc.FO_EDIT)
	field.SetOptionsOff(gc.FO_ACTIVE)
	field.SetOptionsOn(gc.FO_VISIBLE)

	return field, oth
}
