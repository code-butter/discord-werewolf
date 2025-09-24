package models

type Guild struct {
	Id       string `gorm:"primary_key"`
	Name     string
	Channels JsonArray[GuildChannel]
}

type GuildChannel struct {
	Id    string
	Name  string
	AppId string
}
