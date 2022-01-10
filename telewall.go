package main

import (
	"encoding/json"
	"fmt"

	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/godbus/dbus"
	"golang.org/x/net/proxy"
)

/*
appcnf.json config
*/
type SAppConfig struct {
	App []struct {
		DbusZoneName                    string   `json:"DbusZoneName"`
		TelegramBotDenyAccessMessage    string   `json:"TelegramBotDenyAccessMessage"`
		TelegramBotToken                string   `json:"TelegramBotToken"`
		TelegramBotPanicTime            int      `json:"TelegramBotPanicTime"`
		TelegramBotErrorUserInput       string   `json:"TelegramBotErrorUserInput"`
		TelegramBotCommandEnableIP      string   `json:"TelegramBotCommandEnableIP"`
		TelegramBotCommandDisableIP     string   `json:"TelegramBotCommandDisableIP"`
		TelegramBotCommandDisableAll    string   `json:"TelegramBotCommandDisableAll"`
		TelegramBotCommandPanic         string   `json:"TelegramBotCommandPanic"`
		TelegramBotCompleteReloadEnable bool     `json:"TelegramBotCompleteReloadEnable"`
		TelegramBotIPListIsEmpty        string   `json:"TelegramBotIPListIsEmpty"`
		TelegramUserList                []string `json:"TelegramUserList"`
		TelegramBotPanicModeOff         string   `json:"TelegramBotPanicModeOff"`
		TelegramBotDebugMode            bool     `json:"TelegramBotDebugMode"`
		TelegramBotProxyEnable          bool     `json:"TelegramBotProxyEnable"`
		TelegramBotProxyAddressAndPort  string   `json:"TelegramBotProxyAddressAndPort"`
		TelegramBotProxyLogin           string   `json:"TelegramBotProxyLogin"`
		TelegramBotProxyPassword        string   `json:"TelegramBotProxyPassword"`
	} `json:"App"`
}

/*
Panic Function.
*/
func fPanicModeOn(timemin int) {

	log.Println("FUNC PanicMode Activate")
	log.Println("time panic", timemin)

	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		panic(err)
	}

	if err = conn.Auth(nil); err != nil {
		conn.Close()
		conn = nil
	}

	if err = conn.Hello(); err != nil {
		conn.Close()
		conn = nil
	}

	defer conn.Close()

	obj := conn.Object("org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1")

	call := obj.Call("org.fedoraproject.FirewallD1.enablePanicMode", 0)
	if call.Err != nil {
		panic(call.Err)
	}

	log.Println("Panic Mode Enable On Min:", timemin)
	time.Sleep(time.Duration(timemin) * time.Minute)
	log.Println("Panic Mode Disable")

	call = obj.Call("org.fedoraproject.FirewallD1.disablePanicMode", 0)
	if call.Err != nil {
		panic(call.Err)
	}

}

/*
Enable IP
*/
func fEnableIp(ipin string, zonePath string) {

	ipaddr := net.ParseIP(ipin)

	stipaddr := fmt.Sprintf("%s", ipaddr)

	callMethod := "org.fedoraproject.FirewallD1.config.zone.addSource"

	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		panic(err)
	}

	if err = conn.Auth(nil); err != nil {
		conn.Close()
		conn = nil
		return
	}
	if err = conn.Hello(); err != nil {
		conn.Close()
		conn = nil
	}

	defer conn.Close()

	obj := conn.Object("org.fedoraproject.FirewallD1", dbus.ObjectPath(zonePath))

	call := obj.Call(callMethod, 0, stipaddr)
	if call.Err != nil {
		log.Println("ip already present. use S")
	} else {
		log.Println("ip enable - reload firewall")
		obj = conn.Object("org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1")
		call = obj.Call("org.fedoraproject.FirewallD1.reload", 0)
		if call.Err != nil {
			panic(call.Err)
		}
	}

}

/*
Disable IP
*/
func fDisableIp(ipin string, zonePath string, cre bool) {

	ipaddr := net.ParseIP(ipin)

	stipaddr := fmt.Sprintf("%s", ipaddr)

	callMethod := "org.fedoraproject.FirewallD1.config.zone.removeSource"

	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		panic(err)
	}

	if err = conn.Auth(nil); err != nil {
		conn.Close()
		conn = nil
		return
	}
	if err = conn.Hello(); err != nil {
		conn.Close()
		conn = nil
	}

	defer conn.Close()

	obj := conn.Object("org.fedoraproject.FirewallD1", dbus.ObjectPath(zonePath))

	call := obj.Call(callMethod, 0, stipaddr)
	if call.Err != nil {
		log.Println("ip does not exist. use S")
	} else {

		log.Println("ip disable - firewall reload")

		obj = conn.Object("org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1")

		if cre == true {
			log.Println("Complite Reload FireWall")
			call = obj.Call("org.fedoraproject.FirewallD1.completeReload", 0)
			if call.Err != nil {
				panic(call.Err)

			}
		} else {
			call = obj.Call("org.fedoraproject.FirewallD1.reload", 0)
			log.Println("Reload FireWall")
			if call.Err != nil {
				panic(call.Err)
			}
		}
	}

}

/*
Delete all ip in Trusted Zone
*/
func fDisableAllIp(zonePath string, cre bool) {

	callMethod := "org.fedoraproject.FirewallD1.config.zone.getSources"

	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		panic(err)
	}

	if err = conn.Auth(nil); err != nil {
		conn.Close()
		conn = nil
		return
	}
	if err = conn.Hello(); err != nil {
		conn.Close()
		conn = nil
	}

	defer conn.Close()

	obj := conn.Object("org.fedoraproject.FirewallD1", dbus.ObjectPath(zonePath))

	var s []string

	call := obj.Call(callMethod, 0).Store(&s)

	if call != nil {
		fmt.Fprintln(os.Stderr, "Failed to get list of ip:", err)
	}

	countlen := len(s)
	if countlen > 0 {

		callMethod = "org.fedoraproject.FirewallD1.config.zone.removeSource"

		for _, v := range s {
			sv := fmt.Sprintf("%s", v)
			call := obj.Call(callMethod, 0, sv)
			if call.Err != nil {
				log.Println("empty list of ip")
			}
		}

		log.Println("ip disable all - firewall reload")

		obj = conn.Object("org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1")

		if cre == true {
			log.Println("Complete Reload FireWall")
			callreload := obj.Call("org.fedoraproject.FirewallD1.completeReload", 0)
			if callreload.Err != nil {
				panic(callreload.Err)
			}
		} else {
			log.Println("Reload FireWall")
			callreload := obj.Call("org.fedoraproject.FirewallD1.reload", 0)
			if callreload.Err != nil {
				panic(callreload.Err)
			}
		}

	}

}

/*
List IP in Trusted Zone
*/
func fStatusIp(zonePath string) string {

	callMethod := "org.fedoraproject.FirewallD1.config.zone.getSources"

	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		panic(err)
	}

	if err = conn.Auth(nil); err != nil {
		conn.Close()
		conn = nil
		//return
	}
	if err = conn.Hello(); err != nil {
		conn.Close()
		conn = nil
	}

	defer conn.Close()

	obj := conn.Object("org.fedoraproject.FirewallD1", dbus.ObjectPath(zonePath))
	var s []string
	call := obj.Call(callMethod, 0).Store(&s)
	if call != nil {
		fmt.Fprintln(os.Stderr, "Failed to get list of ip:", err)
		os.Exit(1)
	}

	retline := strings.Join(s, " ")

	return retline

}

/*
Command Parse
e - Enable ip
d - Disable ip
l - disable all
s - Status
p - Panic
r - Reload
f - Full Reload
h - Print Help
*/
func fParceCommand(parsestring string) (int, string, bool) {

	comnum := 0
	ipstring := "0"
	errorflag := false
	commandstr := "0"
	cnt := false

	countcommand := len(parsestring)
	if countcommand > 17 {
		errorflag = true
		return comnum, ipstring, errorflag
	}

	prs := strings.ToLower(parsestring)

	eo := strings.HasPrefix(prs, "e ")
	do := strings.HasPrefix(prs, "d ")

	//if the command contains ip
	if eo == true || do == true {
		masstring := strings.Split(prs, " ")
		commandstr, ipstring = masstring[0], masstring[1]
		cnt = true
	}

	//command s, r, f, l, p, h
	if prs == "s" || prs == "r" || prs == "f" || prs == "l" || prs == "p" || prs == "h" {
		commandstr = prs
		cnt = true
	}

	if cnt == false {
		errorflag = true
		return comnum, ipstring, errorflag
	}

	i := commandstr
	switch i {
	case "e":
		comnum = 1
		break
	case "d":
		comnum = 2
		break
	case "l":
		comnum = 3
		break
	case "s":
		comnum = 4
		break
	case "p":
		comnum = 5
		break
	case "r":
		comnum = 6
		break
	case "f":
		comnum = 7
		break
	case "h":
		comnum = 8
	}

	//if the command contains ip
	if comnum == 1 || comnum == 2 {
		checkipvar := nCheckIp(ipstring)
		if checkipvar == false {
			comnum = 0
			ipstring = "0"
			errorflag = true
			return comnum, ipstring, errorflag
		}
	}

	return comnum, ipstring, errorflag
}

/*
Extend ip check.
Prohibits content: "0" in 1 octet (0.*.*.*) or "255" in any octet
*/
func nCheckIp(inputstring string) bool {

	countip := len(inputstring)

	if countip > 15 || countip <= 0 {
		return false
	}

	ipaddr := net.ParseIP(inputstring).To4()

	if ipaddr == nil {
		return false
	}

	if ipaddr[0] <= 0 || ipaddr[0] >= 255 {
		return false
	}

	if ipaddr[1] < 0 || ipaddr[1] >= 255 {
		return false
	}

	if ipaddr[2] < 0 || ipaddr[1] >= 255 {
		return false
	}

	if ipaddr[3] < 0 || ipaddr[3] >= 255 {
		return false
	}

	return true

}

/*
Socks5 connector
*/
func socksProxyClient(address, user, password string) *http.Client {

	socksAuth := proxy.Auth{User: user, Password: password}
	dialSocksProxy, err := proxy.SOCKS5(
		"tcp",
		address,
		&socksAuth,
		proxy.Direct,
	)

	if err != nil {
		fmt.Println("Error connecting to proxy:", err)
	}

	socksClient := &http.Client{Transport: &http.Transport{Dial: dialSocksProxy.Dial}}
	return socksClient
}

/*
Get Zone
*/
func getZoneByName(zoneName string) (zonePath string) {

	callMethod := "org.fedoraproject.FirewallD1.config.getZoneByName"

	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		panic(err)
	}

	if err = conn.Auth(nil); err != nil {
		conn.Close()
		conn = nil

	}
	if err = conn.Hello(); err != nil {
		conn.Close()
		conn = nil
	}

	defer conn.Close()

	obj := conn.Object("org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1/config")

	call := obj.Call(callMethod, 0, zoneName).Store(&zonePath)

	if call != nil {
		fmt.Fprintln(os.Stderr, "Failed to get Zone", err)
		os.Exit(1)
	}

	log.Println("Return Zone Path is:", zonePath)

	return zonePath

}

/*
Reload FireWall
*/
func reloadFireWall(completeReloadEnable bool) {

	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		panic(err)
	}

	if err = conn.Auth(nil); err != nil {
		conn.Close()
		conn = nil

	}
	if err = conn.Hello(); err != nil {
		conn.Close()
		conn = nil
	}

	defer conn.Close()

	obj := conn.Object("org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1")

	if completeReloadEnable == true {
		call := obj.Call("org.fedoraproject.FirewallD1.completeReload", 0)
		if call.Err != nil {
			panic(call.Err)
		}
	} else {
		call := obj.Call("org.fedoraproject.FirewallD1.reload", 0)
		if call.Err != nil {
			panic(call.Err)
		}
	}

}

/*
Check DBUS connection
*/
func init() {

	callMethod := "org.fedoraproject.FirewallD1.getDefaultZone"

	conn, err := dbus.SystemBusPrivate()
	if err != nil {
		panic(err)
	}

	if err = conn.Auth(nil); err != nil {
		conn.Close()
		conn = nil
		//return
	}
	if err = conn.Hello(); err != nil {
		conn.Close()
		conn = nil
	}
	var s string
	obj := conn.Object("org.fedoraproject.FirewallD1", "/org/fedoraproject/FirewallD1")
	call := obj.Call(callMethod, 0).Store(&s)
	log.Println("DefaultZone is", s)

	if call != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to dbus org.fedoraproject.FirewallD1", err)
		os.Exit(100)
	}
	conn.Close()
}

func main() {

	//Read Config
	file, _ := os.Open("appcnf.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := new(SAppConfig)
	err := decoder.Decode(&config)
	if err != nil {
		log.Println("error in decoder")
	}

	log.Println("Telegram Bot Proxy is:", config.App[0].TelegramBotProxyEnable)

	dbusZonePath := getZoneByName(config.App[0].DbusZoneName)

	var bot *tgbotapi.BotAPI

	//change proxy mode
	if config.App[0].TelegramBotProxyEnable == true {
		socksClient := socksProxyClient(config.App[0].TelegramBotProxyAddressAndPort, config.App[0].TelegramBotProxyLogin, config.App[0].TelegramBotProxyPassword)
		bot, err = tgbotapi.NewBotAPIWithClient(config.App[0].TelegramBotToken, "Default01", socksClient)
		if err != nil {
			log.Panic(err)
		}

	} else {

		bot, err = tgbotapi.NewBotAPI(config.App[0].TelegramBotToken)
		if err != nil {
			log.Panic(err)
		}
	}

	bot.Debug = config.App[0].TelegramBotDebugMode
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { //waiting for a message
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		currentUser := update.Message.From.UserName
		accessGrant := false
		log.Println("user for check", currentUser)

		//check user
		for _, au := range config.App[0].TelegramUserList {
			cmpusr := 1
			cmpusr = strings.Compare(au, currentUser)
			if cmpusr == 0 {
				accessGrant = true
				break
			}
		}

		log.Println("For User", currentUser, "Access Grant", accessGrant)
		// accessGrant true access granted

		txttotlg := "0"
		if accessGrant == true {

			textmsg := update.Message.Text

			mcomnum := 0
			mipparse := "0"
			merrorflag := true

			//start command handler
			mcomnum, mipparse, merrorflag = fParceCommand(textmsg)

			if merrorflag == true {
				txttotlg = config.App[0].TelegramBotErrorUserInput

			} else {

				switch mcomnum {

				case 1:
					fEnableIp(mipparse, dbusZonePath)
					txttotlg = config.App[0].TelegramBotCommandEnableIP
					break
				case 2:
					fDisableIp(mipparse, dbusZonePath, config.App[0].TelegramBotCompleteReloadEnable)
					txttotlg = config.App[0].TelegramBotCommandDisableIP
					break
				case 3:
					fDisableAllIp(dbusZonePath, config.App[0].TelegramBotCompleteReloadEnable)
					txttotlg = config.App[0].TelegramBotCommandDisableAll
					break
				case 4:
					txttotlg = fStatusIp(dbusZonePath)
					txttotlglen := len(txttotlg)
					if txttotlglen == 0 {
						txttotlg = config.App[0].TelegramBotIPListIsEmpty
					}
					break
				case 5:
					txttotlg = config.App[0].TelegramBotCommandPanic
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, txttotlg)
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					fPanicModeOn(config.App[0].TelegramBotPanicTime)
					txttotlg = config.App[0].TelegramBotPanicModeOff
					break
				case 6:
					reloadFireWall(false)
					txttotlg = "FireWall Reload"
					break
				case 7:
					txttotlg = "FireWall Full Reload"
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, txttotlg)
					bot.Send(msg)
					reloadFireWall(true)
					break
				case 8:
					txttotlg = "*E* - Enable\n" +
						"*D* - Disable ip\n" +
						"*L* - Disable All\n" +
						"*S* - Status\n" +
						"*P* - Panic\n" +
						"*R* - Soft Firewall Reload\n" +
						"*F* - Full Firewall Reload\n" +
						"*H* - Print This Help\n"
				}
			}
		} else {
			txttotlg = config.App[0].TelegramBotDenyAccessMessage
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, txttotlg)
		msg.ReplyToMessageID = update.Message.MessageID
		msg.ParseMode = "markdown"
		go bot.Send(msg)
	}

}
