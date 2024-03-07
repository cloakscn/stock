package layouts

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

func initTabs(w fyne.Window) ([]*container.TabItem, error) {
	tabIndex, err := getTabIndex(w)
	if err != nil {
		return nil, err
	}
	return []*container.TabItem{
		tabIndex,
		getTabSetting(),
	}, nil
}
