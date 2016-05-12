package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	//	"fmt"
	"encoding/hex"
	"io"
	"log"
	"net"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	mathRand "math/rand"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	mathRand.Seed(int64(time.Now().Nanosecond()))
}

func Rand() int64 {
	return mathRand.Int63()
}

func RandRange(min, max int64) int64 {
	return mathRand.Int63n(max-min) + min
}

func AuthCode() string {
	str := fmt.Sprintf("%06d", mathRand.Int31n(1000000))
	return str
}

func UUID() string {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "UUID-ERROR"
	}
	return Md5Hash(base64.URLEncoding.EncodeToString(b))
}

func RandNickname() string {
	return fmt.Sprintf("nick-%d", time.Now().UnixNano())
}

func Md5Hash(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
	//return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

func CheckError(err error) {
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
}

func GetAppName() string {
	execfile := os.Args[0]
	if runtime.GOOS == `windows` {
		execfile = strings.Replace(execfile, "\\", "/", -1)
	}
	_, filename := path.Split(execfile)
	return filename
}

func GetCurrentPath() string {
	curpath, _ := os.Getwd()
	return curpath
}

func GetIntranetIP() string {
	addrs, err := net.InterfaceAddrs()
	CheckError(err)
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || ip.IsLoopback() {
			continue
		}
		ip = ip.To4()
		if ip == nil {
			continue
		}
		if IsIntranetIP(ip.String()) {
			return ip.String()
		}
	}
	return "127.0.0.1"
}

// 10.0.0.0 ~ 10.255.255.255(A)
// 172.16.0.0 ~ 172.31.255.255(B)
// 192.168.0.0 ~ 192.168.255.255(C)
func IsIntranetIP(ip string) bool {
	if strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "192.168.") {
		return true
	}
	if strings.HasPrefix(ip, "172.") {
		arr := strings.Split(ip, ".")
		if len(arr) != 4 {
			return false
		}
		second, err := strconv.ParseInt(arr[1], 10, 64)
		if err != nil {
			return false
		}
		if second >= 16 && second <= 31 {
			return true
		}
	}
	return false
}

func CheckPhone(phone string) bool {
	if len(phone) <= 0 {
		return false
	}
	if m, _ := regexp.MatchString(`^(1[3|4|5|8][0-9]\d{4,8})$`, phone); !m {
		return false
	}
	return true
}

func ChechEmail(email string) bool {
	if len(email) <= 0 {
		return false
	}
	if m, _ := regexp.MatchString(`^([a-zA-Z0-9_\.\-])+\@(([a-zA-Z0-9\-])+\.)+([a-zA-Z0-9]{2,4})+$`, email); !m {
		return false
	}
	return true
}
