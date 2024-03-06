package main

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/go-resty/resty/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	executablePath string
	db             *gorm.DB
)

type response struct {
	Dlmkts string
	Full   int64
	Lt     int64
	Rc     int64
	Rt     int64
	Svr    int64
	Data   data
}

type data struct {
	Total int64
	Diff  []diff
}

type diff struct {
	F1   int64
	F2   int64
	F3   int64
	F4   int64
	F6   float64
	F12  string
	F13  int64
	F14  string
	F104 interface{}
	F105 interface{}
	F106 interface{}
	F152 int64
}

func init() {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}
	executablePath = filepath.Dir(executable)

	err = initDatabase()
	if err != nil {
		panic(err)
	}
}

func main() {
	a := app.New()
	a.SetIcon(resourceKPng)
	a.Settings().SetTheme(&Theme{})

	dashboard, err := initDashboard(a)
	if err != nil {
		panic(err)
	}
	defer dashboard.ShowAndRun()

	initTray(a, dashboard)
	initLifecycle(a)
}

type Code struct {
	gorm.Model
	Code string
}

func initDatabase() (err error) {
	db, err = gorm.Open(sqlite.Open(filepath.Join(executablePath, "cache.db")), &gorm.Config{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&Code{})
	if err != nil {
		return err
	}

	return nil
}

func initDashboard(a fyne.App) (fyne.Window, error) {
	dashboard := a.NewWindow("Stock")
	dashboard.SetMainMenu(makeMenu(a, dashboard))
	dashboard.SetMaster()
	dashboard.Resize(fyne.NewSize(640, 460))

	tabItems, err := initTabs(dashboard)
	if err != nil {
		return nil, err
	}
	tabs := container.NewAppTabs(tabItems...)
	tabs.SetTabLocation(container.TabLocationLeading)
	dashboard.SetCloseIntercept(func() { dashboard.Hide() })
	dashboard.SetContent(tabs)
	return dashboard, nil
}

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

//func getTreeTab() {
//	content := container.NewStack()
//	title := widget.NewLabel("Component name")
//	intro := widget.NewLabel("An introduction would probably go\nhere, as well as a")
//	intro.Wrapping = fyne.TextWrapWord
//
//	tutorial := container.NewBorder(
//		container.NewVBox(title, widget.NewSeparator(), intro),
//		nil, nil, nil, content)
//
//	if fyne.CurrentDevice().IsMobile() {
//		// 适配移动端
//		//dashboard.SetContent(makeNav(setTutorial, false))
//	} else {
//		split := container.NewHSplit(widget.NewLabel("hello"), tutorial)
//		split.Offset = 0.2
//		dashboard.SetContent(split)
//	}
//
//
//}

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

func makeMenu(a fyne.App, w fyne.Window) *fyne.MainMenu {
	newItem := fyne.NewMenuItem("New", nil)
	checkedItem := fyne.NewMenuItem("Checked", nil)
	checkedItem.Checked = true
	disabledItem := fyne.NewMenuItem("Disabled", nil)
	disabledItem.Disabled = true
	otherItem := fyne.NewMenuItem("Other", nil)
	mailItem := fyne.NewMenuItem("Mail", func() { fmt.Println("Menu New->Other->Mail") })
	mailItem.Icon = theme.MailComposeIcon()
	otherItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Project", func() { fmt.Println("Menu New->Other->Project") }),
		mailItem,
	)
	fileItem := fyne.NewMenuItem("File", func() { fmt.Println("Menu New->File") })
	fileItem.Icon = theme.FileIcon()
	dirItem := fyne.NewMenuItem("Directory", func() { fmt.Println("Menu New->Directory") })
	dirItem.Icon = theme.FolderIcon()
	newItem.ChildMenu = fyne.NewMenu("",
		fileItem,
		dirItem,
		otherItem,
	)

	openSettings := func() {
		w := a.NewWindow("Fyne Settings")
		w.SetContent(settings.NewSettings().LoadAppearanceScreen(w))
		w.Resize(fyne.NewSize(440, 520))
		w.Show()
	}
	settingsItem := fyne.NewMenuItem("Settings", openSettings)
	settingsShortcut := &desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: fyne.KeyModifierShortcutDefault}
	settingsItem.Shortcut = settingsShortcut
	w.Canvas().AddShortcut(settingsShortcut, func(shortcut fyne.Shortcut) {
		openSettings()
	})

	cutShortcut := &fyne.ShortcutCut{Clipboard: w.Clipboard()}
	cutItem := fyne.NewMenuItem("Cut", func() {
		shortcutFocused(cutShortcut, w)
	})
	cutItem.Shortcut = cutShortcut
	copyShortcut := &fyne.ShortcutCopy{Clipboard: w.Clipboard()}
	copyItem := fyne.NewMenuItem("Copy", func() {
		shortcutFocused(copyShortcut, w)
	})
	copyItem.Shortcut = copyShortcut
	pasteShortcut := &fyne.ShortcutPaste{Clipboard: w.Clipboard()}
	pasteItem := fyne.NewMenuItem("Paste", func() {
		shortcutFocused(pasteShortcut, w)
	})
	pasteItem.Shortcut = pasteShortcut
	performFind := func() { fmt.Println("Menu Find") }
	findItem := fyne.NewMenuItem("Find", performFind)
	findItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyF, Modifier: fyne.KeyModifierShortcutDefault | fyne.KeyModifierAlt | fyne.KeyModifierShift | fyne.KeyModifierControl | fyne.KeyModifierSuper}
	w.Canvas().AddShortcut(findItem.Shortcut, func(shortcut fyne.Shortcut) {
		performFind()
	})

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("Documentation", func() {
			u, _ := url.Parse("https://developer.fyne.io")
			_ = a.OpenURL(u)
		}),
		fyne.NewMenuItem("Support", func() {
			u, _ := url.Parse("https://fyne.io/support/")
			_ = a.OpenURL(u)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Sponsor", func() {
			u, _ := url.Parse("https://fyne.io/sponsor/")
			_ = a.OpenURL(u)
		}))

	// a quit item will be appended to our first (File) menu
	file := fyne.NewMenu("File", newItem, checkedItem, disabledItem)
	device := fyne.CurrentDevice()
	if !device.IsMobile() && !device.IsBrowser() {
		file.Items = append(file.Items, fyne.NewMenuItemSeparator(), settingsItem)
	}
	main := fyne.NewMainMenu(
		file,
		fyne.NewMenu("Edit", cutItem, copyItem, pasteItem, fyne.NewMenuItemSeparator(), findItem),
		helpMenu,
	)
	checkedItem.Action = func() {
		checkedItem.Checked = !checkedItem.Checked
		main.Refresh()
	}
	return main
}

func shortcutFocused(s fyne.Shortcut, w fyne.Window) {
	switch sh := s.(type) {
	case *fyne.ShortcutCopy:
		sh.Clipboard = w.Clipboard()
	case *fyne.ShortcutCut:
		sh.Clipboard = w.Clipboard()
	case *fyne.ShortcutPaste:
		sh.Clipboard = w.Clipboard()
	}
	if focused, ok := w.Canvas().Focused().(fyne.Shortcutable); ok {
		focused.TypedShortcut(s)
	}
}

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

func selectCodeList() ([]Code, error) {
	var data []Code
	err := db.Model(&Code{}).Find(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func getTabIndex(w fyne.Window) (*container.TabItem, error) {
	latest := widget.NewLabel("-")
	label := widget.NewLabel("获取数据中...")

	codeList, err := selectCodeList()
	if err != nil {
		return nil, err
	}

	var codes []string
	for _, item := range codeList {
		codes = append(codes, item.Code)
	}
	//codes := []string{"1.000001", "0.399001", "1.601099", "1.600280", "0.002616"}

	go func() {
		for range time.Tick(3 * time.Second) {
			if len(codes) == 0 {
				label.SetText("没有 code 信息.")
			}
			resp, err := getData(codes...)
			if err != nil {
				fmt.Println(err)
			} else {
				latest.SetText("【Latest Update】" + time.Now().Format(time.DateTime))
				var content string
				for _, item := range resp.Data.Diff {
					content += fmt.Sprintf("【%s】%s: %.4f\n", item.F12, item.F14, float64(item.F2)/math.Pow(10, float64(item.F1)))
				}
				label.SetText(content)
			}
		}
	}()

	top := container.NewHBox(widget.NewButton("Add New Index", func() {
		code := widget.NewEntry()
		formDialog := dialog.NewForm("Add New Index", "Confirm", "Dismiss",
			[]*widget.FormItem{
				widget.NewFormItem("Code【沪市 1；深市 0】", code),
			},
			func(b bool) {
				if !b {
					return
				}

				err = db.Create(&Code{Code: code.Text}).Error
				if err != nil {
					fmt.Println(err)
				} else {
					codes = append(codes, code.Text)
				}
			}, w)
		formDialog.Show()
	}))

	border := container.NewBorder(top, latest, nil, nil, label)

	return container.NewTabItem("Index", border), nil
}

func getData(codes ...string) (*response, error) {
	now := time.Now().Add(-time.Second).UnixMilli()
	resp, err := resty.New().R().SetQueryParams(map[string]string{
		"fltt":   "1",
		"invt":   "2",
		"cb":     fmt.Sprintf("jQuery35106585774236178676_%d", now),
		"fields": "f12,f13,f14,f1,f2,f4,f3,f152,f6,f104,f105,f106",
		"secids": strings.Join(codes, ","),
		"ut":     "fa5fd1943c7b386f172d6893dbfba10b",
		"pn":     "1",
		"np":     "1",
		"pz":     "20",
		"wbp2u":  "7448496871626718|0|1|0|web",
		"_":      fmt.Sprintf("%d", now),
	}).Get("https://push2.eastmoney.com/api/qt/ulist/get")
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`jQuery\d+_\d+\((.*)\);`)
	matches := re.FindStringSubmatch(string(resp.Body()))

	if len(matches) < 2 {
		return nil, err
	}
	var body = new(response)
	err = json.Unmarshal([]byte(matches[1]), body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
