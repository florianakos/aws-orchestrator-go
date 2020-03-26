
package main

import ui "github.com/gizak/termui"

func main() {
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	data := []int{int(2),
       int(6),
        int(6),
        int(6),
        int(6),
        100,
        int(6),
        int(6),
        int(6),
        int(6),
        100,
        int(6),
        int(6),
        int(6),
        int(6),
        100,
        int(6),
        int(6),
        int(6),
        int(6),
        100,
        int(6),
        int(6),
        int(6),
        int(6),
        100,
        int(6),
        int(6),
        int(6),
        int(6),
        100,
        }

	spl3 := ui.NewSparkline()
	spl3.Data = data
	spl3.Title = ""
	spl3.Height = 8
	spl3.LineColor = ui.ColorYellow


	spls2 := ui.NewSparklines(spl3)
	spls2.Height = 11
	spls2.Width = len(data)+3
	spls2.BorderFg = ui.ColorCyan
	spls2.X = 1
	spls2.Y = 1
	spls2.BorderLabel = "CPU Utilization"

	spls3 := ui.NewSparklines(spl3)
	spls3.Height = 11
	spls3.Width = len(data)+3
	spls3.BorderFg = ui.ColorCyan
	spls3.X = 1
	spls3.Y = 12	
	spls3.BorderLabel = "CPU Utilization"

	ui.Render(spls2, spls3)

	ui.Handle("q", func(ui.Event) {
		ui.StopLoop()
	})
	ui.	Loop()
}