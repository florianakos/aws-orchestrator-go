package main

import ui "github.com/gizak/termui/v3"
import "github.com/gizak/termui/v3/widgets"

func main() {
    if err:=ui.Init(); err != nil {
        panic(err)
    }
    defer ui.Close()

    g0 := widgets.NewGauge()
    g0.Title = "Slim Gauge"
	  g0.SetRect(20, 20, 30, 30)
	  g0.Percent = 75
	  g0.BarColor = ui.ColorRed
	  g0.BorderStyle.Fg = ui.ColorWhite
	  g0.TitleStyle.Fg = ui.ColorCyan

    ui.Render(g0)

  	uiEvents := ui.PollEvents()
  	for {
  		e := <-uiEvents
  		switch e.ID {
  		case "q", "<C-c>":
  			return
  		}
  	}
}
