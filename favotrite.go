package main

import(
  "fmt"
  "math"
  "net/http"
)

func FavoriteShop(w http.ResponseWriter, r *http.Request) {
}













func CalculateSortWeight(text string) string {

	textTest := ""
	var vowelValue float64
	characterCounter := 0.0
	vowelCounter := 0.0
	for _, v := range text {
		if Alphabet[string(v)] == 0 {
			continue
		}

		if Alphabet[string(v)] >= 100 {
			alphabetValue := Alphabet[string(v)]
			vowelCounter = 0 //If found character reset vowelCounter
			if characterCounter == 0.0 {
				textTest = fmt.Sprint(alphabetValue+vowelValue, ".")
			} else {
				textTest += fmt.Sprint(alphabetValue + vowelValue)
			}
			vowelValue = 0
			characterCounter++
		} else {
			temp := int(vowelValue + (Alphabet[string(v)] / (math.Pow(10, vowelCounter))))
			vowelValue = float64(temp)
			vowelCounter++
		}
	}
	return textTest
}

func InitData() map[string]float64 {
	Alphabet := make(map[string]float64)
	Alphabet["เ"] = 5
	Alphabet["แ"] = 6
	Alphabet["โ"] = 7
	Alphabet["ใ"] = 8
	Alphabet["ไ"] = 9
	Alphabet["ะ"] = 10
	Alphabet["ั"] = 11
	Alphabet["า"] = 12
	Alphabet["ำ"] = 13
	Alphabet["ิ"] = 14
	Alphabet["ี"] = 15
	Alphabet["ึ"] = 16
	Alphabet["ื"] = 17
	Alphabet["ุ"] = 18
	Alphabet["ู"] = 19

	Alphabet["็"] = 20
	Alphabet["์"] = 21
	Alphabet["่"] = 22
	Alphabet["้"] = 23
	Alphabet["๊"] = 24
	Alphabet["๋"] = 25

	Alphabet["ก"] = 1000
	Alphabet["ข"] = 1030
	Alphabet["ฃ"] = 1060
	Alphabet["ค"] = 1090
	Alphabet["ฅ"] = 1120
	Alphabet["ฆ"] = 1150
	Alphabet["ง"] = 1180
	Alphabet["จ"] = 1210
	Alphabet["ฉ"] = 1240
	Alphabet["ช"] = 1270
	Alphabet["ซ"] = 1300
	Alphabet["ฌ"] = 1330
	Alphabet["ญ"] = 1360
	Alphabet["ฎ"] = 1390
	Alphabet["ฏ"] = 1420
	Alphabet["ฐ"] = 1450
	Alphabet["ฑ"] = 1480
	Alphabet["ฒ"] = 1510
	Alphabet["ณ"] = 1540
	Alphabet["ด"] = 1570
	Alphabet["ต"] = 1600
	Alphabet["ถ"] = 1630
	Alphabet["ท"] = 1660
	Alphabet["ธ"] = 1690
	Alphabet["น"] = 1720
	Alphabet["บ"] = 1750
	Alphabet["ป"] = 1780
	Alphabet["ผ"] = 1810
	Alphabet["ฝ"] = 1840
	Alphabet["พ"] = 1870
	Alphabet["ฟ"] = 1900
	Alphabet["ภ"] = 1930
	Alphabet["ม"] = 1960
	Alphabet["ย"] = 1990
	Alphabet["ร"] = 2020
	Alphabet["ฤ"] = 2050
	Alphabet["ล"] = 2080
	Alphabet["ฦ"] = 2110
	Alphabet["ว"] = 2140
	Alphabet["ศ"] = 2170
	Alphabet["ษ"] = 2200
	Alphabet["ส"] = 2230
	Alphabet["ห"] = 2260
	Alphabet["ฬ"] = 2290
	Alphabet["อ"] = 2320
	Alphabet["ฮ"] = 2350

	Alphabet["a"] = 2380
	Alphabet["A"] = 2410
	Alphabet["b"] = 2440
	Alphabet["B"] = 2470
	Alphabet["c"] = 2500
	Alphabet["C"] = 2530
	Alphabet["d"] = 2560
	Alphabet["D"] = 2590
	Alphabet["e"] = 2620
	Alphabet["E"] = 2650
	Alphabet["f"] = 2680
	Alphabet["F"] = 2710
	Alphabet["g"] = 2740
	Alphabet["G"] = 2770
	Alphabet["h"] = 2800
	Alphabet["H"] = 2830
	Alphabet["i"] = 2860
	Alphabet["I"] = 2890
	Alphabet["j"] = 2920
	Alphabet["J"] = 2950
	Alphabet["k"] = 2980
	Alphabet["K"] = 3010
	Alphabet["l"] = 3040
	Alphabet["L"] = 3070
	Alphabet["m"] = 3100
	Alphabet["M"] = 3130
	Alphabet["n"] = 3160
	Alphabet["N"] = 3190
	Alphabet["o"] = 3220
	Alphabet["O"] = 3250
	Alphabet["p"] = 3280
	Alphabet["P"] = 3310
	Alphabet["q"] = 3340
	Alphabet["Q"] = 3370
	Alphabet["r"] = 3400
	Alphabet["R"] = 3430
	Alphabet["s"] = 3460
	Alphabet["S"] = 3490
	Alphabet["t"] = 3520
	Alphabet["T"] = 3550
	Alphabet["u"] = 3580
	Alphabet["U"] = 3610
	Alphabet["v"] = 3640
	Alphabet["V"] = 3670
	Alphabet["w"] = 3700
	Alphabet["W"] = 3730
	Alphabet["x"] = 3760
	Alphabet["X"] = 3790
	Alphabet["y"] = 3820
	Alphabet["Y"] = 3850
	Alphabet["z"] = 3880
	Alphabet["Z"] = 3910

	Alphabet["๐"] = 3940
	Alphabet["๑"] = 3970
	Alphabet["๒"] = 4000
	Alphabet["๓"] = 4030
	Alphabet["๔"] = 4060
	Alphabet["๕"] = 4090
	Alphabet["๖"] = 4120
	Alphabet["๗"] = 4150
	Alphabet["๘"] = 4180
	Alphabet["๙"] = 4210

	Alphabet["0"] = 4240
	Alphabet["1"] = 4270
	Alphabet["2"] = 4300
	Alphabet["3"] = 4330
	Alphabet["4"] = 4360
	Alphabet["5"] = 4390
	Alphabet["6"] = 4420
	Alphabet["7"] = 4450
	Alphabet["8"] = 4480
	Alphabet["9"] = 4510

	Alphabet[""] = 4540
	Alphabet["!"] = 4570
	Alphabet["\""] = 4600
	Alphabet["#"] = 4630
	Alphabet["$"] = 4660
	Alphabet["%"] = 4690
	Alphabet["&"] = 4720
	Alphabet["'"] = 4750
	Alphabet["("] = 4780
	Alphabet[")"] = 4810
	Alphabet["*"] = 4840
	Alphabet["+"] = 4870
	Alphabet[","] = 4900
	Alphabet["-"] = 4930
	Alphabet["."] = 4960
	Alphabet["/"] = 4990

	Alphabet[":"] = 5020
	Alphabet[";"] = 5050
	Alphabet["<"] = 5080
	Alphabet["="] = 5110
	Alphabet[">"] = 5140
	Alphabet["?"] = 5170
	Alphabet["@"] = 5200

	Alphabet["["] = 5230
	Alphabet["\\"] = 5260
	Alphabet["]"] = 5290
	Alphabet["^"] = 5320
	Alphabet["_"] = 5350
	Alphabet["`"] = 5380

	Alphabet["{"] = 5410
	Alphabet["|"] = 5440
	Alphabet["}"] = 5470
	Alphabet["~"] = 5500

	return Alphabet
}
