package fav

import (
	"time"
)

//Map data
var Alphabet map[string]float64

type FavRequest struct {
	MobileNo   string `json:"MobileNo" bson:"MobileNo"`
	MID        string `json:"MID" bson:"MID"`
	IsFavorite bool   `json:"is_favorite" bson:"is_favorite"`
}

type Merchants struct {
	MID          string    `json:"mid" bson:"mid"`
	Name         string    `json:"name" bson:"name"`
	Image        string    `json:"image" bson:"image"`
	Weight       string    `json:"weight" bson:"weight"`
	MainCategory string    `json:"main_category" bson:"main_category"`
	DateCreate   time.Time `json:"date_create" bson:"date_create"`
}

type FavoriteShops struct {
	TelNo     string      `bson:"tel_no" json:"tel_no"`
	Merchants []Merchants `bson:"merchants" json:"merchants"`
}
