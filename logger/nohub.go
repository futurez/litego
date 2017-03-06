package logger

//	"os"

type Alert interface {
	Alert(appname string)
}

var alert Alert

func backupNohup() {
	//	fd, err := os.OpenFile("nohub.out", os.O_RDONLY, 0666)
	//	if err != nil {
	//		return
	//	}

	//	ret, err := fd.Seek(0, os.SEEK_END)
	//	if err != nil {
	//		os.Rename("nohub.out", "nohub.bak")
	//		if alert != nil {
	//			alert.Alert("GetAppName() restart!")
	//		}
	//		return
	//	}

}
