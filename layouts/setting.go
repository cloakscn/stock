package layouts

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func getTabSetting() *container.TabItem {
	language := widget.NewSelect([]string{
		"English",
		"简体中文",
		"繁体中文",
	}, func(s string) {

	})
	language.SetSelectedIndex(0)

	form := widget.NewForm(
		widget.NewFormItem("Language", language),
		//widget.NewFormItem("Language", nil),
	)
	return container.NewTabItem("Settings", form)
}
