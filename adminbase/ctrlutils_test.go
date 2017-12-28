package adminbase

import (
	"testing"
	"time"
)

func TestPascalToSnake(t *testing.T) {
	if orig := "AppId"; pascalToSnake(orig) != "app_id" {
		t.Error("pascalToSnake error", orig, pascalToSnake(orig))
	}
	if orig := "AppID"; pascalToSnake(orig) != "app_id" {
		t.Error("pascalToSnake error", orig, pascalToSnake(orig))
	}
	if orig := "ID"; pascalToSnake(orig) != "id" {
		t.Error("pascalToSnake error", orig, pascalToSnake(orig))
	}
	if orig := "IPAddr"; pascalToSnake(orig) != "ip_addr" {
		t.Error("pascalToSnake error", orig, pascalToSnake(orig))
	}
	if orig := "Name"; pascalToSnake(orig) != "name" {
		t.Error("pascalToSnake error", orig, pascalToSnake(orig))
	}
	if orig := "Name1"; pascalToSnake(orig) != "name_1" {
		t.Error("pascalToSnake error", orig, pascalToSnake(orig))
	}
	if orig := "UserName1984"; pascalToSnake(orig) != "user_name_1984" {
		t.Error("pascalToSnake error", orig, pascalToSnake(orig))
	}
}

type account struct {
	ID             int64     `xorm:"'id' notnull pk autoincr"`
	Username       string    `xorm:"'username' varchar(15)"`
	Password       string    `xorm:"'password' varchar(60)"`
	MachineID      string    `xorm:"'machine_id' varchar(40)"`
	Status         int       `xorm:"'status' notnull"`
	Coins          int64     `xorm:"'coins'"`
	Diamond        int       `xorm:"'diamond' notnull"`
	Money          int       `xorm:"'money' notnull"`
	Nickname       string    `xorm:"'nickname' varchar(50) notnull"`
	PhoneNumber    string    `xorm:"'phone_number' varchar(20)"`
	Icon           string    `xorm:"'icon' varchar(100) notnull"`
	Gender         int       `xorm:"'gender' notnull"`
	RegistrationID string    `xorm:"'registration_id' varchar(40) notnull"`
	ChannelName    string    `xorm:"'channel_name' varchar(128) notnull"`
	SourceChannel  string    `xorm:"'source_channel' varchar(128) notnull"`
	Created        time.Time `xorm:"'created' DATETIME notnull"`
	Updated        time.Time `xorm:"'updated' DATETIME notnull"`
}

func TestMakeSnakeRenderers(t *testing.T) {
	m := new(account)
	r := MakeSnakeRenderers(m, "Password")
	if len(r) != 16 {
		t.Error("len error", r)
	}
}

func TestMakeRenderersByRules(t *testing.T) {
	r := MakeRenderersByRules(`
		Name
		Name: name123
		Hello: world

	`)
	if len(r) != 3 {
		t.Error("len error", r)
	}
}
