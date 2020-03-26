package main

import ui "github.com/gizak/termui"

func main() {
    if err:=ui.Init(); err != nil {
        panic(err)
    }
    defer ui.Close()

    g := ui.NewGauge()
    g.Percent = 50
    g.Width = 50
    g.BorderLabel = "Gauge"

 

    ui.Body.AddRows(ui.NewRow(ui.NewCol(6, 0, g)))

    // calculate layout
    ui.Body.Align()

    ui.Render(ui.Body)

    ui.Loop()
}