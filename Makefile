dist:
	fyne-cross linux -arch=* -env='GOPROXY=https://goproxy.cn,direct' -icon=./resources/static/Icon.png
	fyne-cross windows -app-id=com.fyne-stock -arch=* -env='GOPROXY=https://goproxy.cn,direct' -icon=./resources/static/Icon.png
