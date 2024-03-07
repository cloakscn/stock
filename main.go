package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"github.com/cloakscn/fyne-stock/layouts"
	"log"
)

func main() {
	a := app.New()
	a.SetIcon(resourceKPng)
	a.Settings().SetTheme(&Theme{})

	dashboard, err := layouts.InitDashboard(a)
	if err != nil {
		panic(err)
	}
	defer dashboard.ShowAndRun()

	initTray(a, dashboard)
	initLifecycle(a)
}

func initLifecycle(a fyne.App) {
	a.Lifecycle().SetOnStarted(func() {
		log.Println("Lifecycle: Started")
	})
	a.Lifecycle().SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
	})
	a.Lifecycle().SetOnEnteredForeground(func() {
		log.Println("Lifecycle: Entered Foreground")
	})
	a.Lifecycle().SetOnExitedForeground(func() {
		log.Println("Lifecycle: Exited Foreground")
	})
}

func initTray(a fyne.App, w fyne.Window) {
	if desk, ok := a.(desktop.App); ok {
		menu := fyne.NewMenu("Hello World")
		defer desk.SetSystemTrayMenu(menu)

		dashboard := fyne.NewMenuItem("Dashboard", func() {
			w.Show()
		})

		help := fyne.NewMenuItem("Hello", func() {})
		help.Icon = theme.HomeIcon()
		help.Action = func() {
			log.Println("System tray menu tapped")
			menu.Refresh()
		}

		menu.Items = append(menu.Items, dashboard, help)
	}
}
