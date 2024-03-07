package layouts

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/cloakscn/fyne-stock/constants"
	"github.com/cloakscn/fyne-stock/model"
	"github.com/go-resty/resty/v2"
	"math"
	"regexp"
	"strings"
	"time"
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

func getTabIndex(w fyne.Window) (*container.TabItem, error) {
	notice := widget.NewLabel("Fetch data...")

	codeList, err := selectCodeList()
	if err != nil {
		return nil, err
	}

	var codes []string
	for _, item := range codeList {
		codes = append(codes, item.Code)
	}

	var res *response
	for {
		res, err = getData(codes...)
		if err == nil {
			break
		}
		fmt.Println(err)
	}

	cells := createTableBindingCell(int(res.Data.Total+1), "Code", "Name", "Price", "Stage", "Remark")

	table := widget.NewTable(func() (rows int, cols int) {
		return len(cells), len(cells[0])
	}, func() fyne.CanvasObject {
		return widget.NewLabel("")
	}, func(id widget.TableCellID, object fyne.CanvasObject) {
		label := object.(*widget.Label)
		if id.Row == 0 {
			label.SetText(cells[0][id.Col].(string))
		} else {
			label.Bind(cells[id.Row][id.Col].(binding.String))
		}
	})

	table.SetColumnWidth(0, 60)
	table.SetColumnWidth(1, 80)
	table.SetColumnWidth(2, 120)
	table.SetColumnWidth(3, 120)
	table.SetColumnWidth(4, 120)

	table.OnSelected = func(id widget.TableCellID) {
		if id.Row == 0 {
			return
		}

		_stage, err := cells[id.Row][3].(binding.String).Get()
		if err != nil {
			fmt.Println(err)
		}
		stage := widget.NewEntry()
		stage.SetText(_stage)

		_remark, err := cells[id.Row][4].(binding.String).Get()
		if err != nil {
			fmt.Println(err)
		}
		remark := widget.NewEntry()
		remark.SetText(_remark)

		formDialog := dialog.NewForm("Modify "+cells[0][id.Col].(string), "Confirm", "Dismiss",
			[]*widget.FormItem{
				widget.NewFormItem(cells[0][0].(string), widget.NewLabelWithData(cells[id.Row][0].(binding.String))),
				widget.NewFormItem(cells[0][1].(string), widget.NewLabelWithData(cells[id.Row][1].(binding.String))),
				widget.NewFormItem(cells[0][2].(string), widget.NewLabelWithData(cells[id.Row][2].(binding.String))),
				widget.NewFormItem(cells[0][3].(string), stage),
				widget.NewFormItem(cells[0][4].(string), remark),
			}, func(b bool) {
				if !b {
					return
				}

				var update = make(map[string]interface{})
				update["stage"] = stage.Text
				update["remark"] = remark.Text

				_code, err := cells[id.Row][0].(binding.String).Get()
				if err != nil {
					fmt.Println(err)
					dialog.NewError(err, w).Show()
					return
				}
				err = constants.DB().Model(&model.Code{}).Where("code like ?", "%"+_code).Updates(update).Error
				if err != nil {
					fmt.Println(err)
				}
			}, w)
		formDialog.Resize(fyne.NewSize(400, -1))
		formDialog.Show()
	}

	go func() {
		for range time.Tick(3 * time.Second) {
			if len(codes) == 0 {
				notice.SetText("Not found code info.")
			} else {
				res, err = getData(codes...)
				if err != nil {
					fmt.Println(err)
					continue
				}
			}

			notice.SetText("【Latest Update】" + time.Now().Format(time.DateTime))

			for i, item := range res.Data.Diff {
				cells[i+1][0].(binding.String).Set(item.F12)
				cells[i+1][1].(binding.String).Set(item.F14)
				cells[i+1][2].(binding.String).Set(fmt.Sprintf("%.4f", float64(item.F2)/math.Pow(10, float64(item.F1))))

				var code model.Code
				err = constants.DB().Model(&model.Code{}).Where("code like ?", "%"+item.F12).First(&code).Error
				if err != nil {
					continue
				}
				cells[i+1][3].(binding.String).Set(code.Stage)
				cells[i+1][4].(binding.String).Set(code.Remark)
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

				err = constants.DB().Create(&model.Code{Code: code.Text}).Error
				if err != nil {
					fmt.Println(err)
				} else {
					codes = append(codes, code.Text)
				}
				appendTableBindingCell(cells, 1)
			}, w)
		formDialog.Resize(fyne.NewSize(400, -1))
		formDialog.Show()
	}))
	border := container.NewBorder(top, notice, nil, nil, table)
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

func selectCodeList() ([]model.Code, error) {
	var data []model.Code
	err := constants.DB().Model(&model.Code{}).Find(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func createTableBindingCell(raw int, headers ...string) [][]interface{} {
	cells := make([][]interface{}, raw+1)
	for _, header := range headers {
		cells[0] = append(cells[0], header)
	}
	for i := 0; i < raw; i++ {
		for _, _ = range headers {
			cells[i+1] = append(cells[i+1], binding.NewString())
		}
	}
	return cells
}

func appendTableBindingCell(cells [][]interface{}, count int) [][]interface{} {
	for i := 0; i < count; i++ {
		var cols []interface{}
		for _, _ = range cells[0] {
			cols = append(cols, binding.NewString())
		}
		cells = append(cells, cols)
	}
	return cells
}
