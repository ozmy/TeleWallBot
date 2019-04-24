#**TeleWall**

#About
    
Project to control firewalld via Telegram.
Allows you to add and remove ip to the trusted zone firewalld.
Сan make a completely invisible server.
Socks5 Proxy Support.

#System requirements

Need support in the operating system Linux - Firewalld, DBUS
If you get error 100 you need to install DBUS.
    
# Configuration

Config File -  appcnf.json

# Instalation

1. go get github.com/ozmy/telewall
2. go build github.com/ozmy/telewall
3. Change config appcnf.json, in TelegramUserList use UserName in Telegram without "@"
4. Copy bin file and appcnf.json to /opt/telewall or change WorkingDirectory settings in telewall.service file
5. Copy file telewall.service to /etc/systemd/system/
6. systemctl enable telewall
7. systemctl start telewall
8. give the command to bot


# Bot Commands
(no case sensetive)
1. E 192.168.1.1 - enable ip
2. D 192.168.1.1 - disable ip
3. L             - disable all ip
4. S             - Status. List ip address in Trusted Zone
5. p             - Panic Mode On (default panic 1 min)
6. R             - Reload Firewall
7. F             - Full Reload Firewall


# License

You may copy, distribute and modify the software provided that modifications are described and licensed for free under LGPL-3. Derivatives works (including modifications or anything statically linked to the library) can only be redistributed under LGPL-3, but applications that use the library don't have to be.