package mockdata

type MerchantMaster struct {
	ID              string
	Name            string
	HotelCount      int
	PrefectureCodes []string
}

func MerchantMasters() []MerchantMaster {

	all :=
		AllPrefectureCodes()

	hokkaido :=
		[]string{
			"01",
		}

	okinawa :=
		[]string{
			"47",
		}

	tohoku :=
		[]string{
			"02",
			"03",
			"04",
			"05",
			"06",
			"07",
		}

	kanto :=
		[]string{
			"08",
			"09",
			"10",
			"11",
			"12",
			"13",
			"14",
		}

	tokai :=
		[]string{
			"21",
			"22",
			"23",
			"24",
		}

	kinki :=
		[]string{
			"25",
			"26",
			"27",
			"28",
			"29",
			"30",
		}

	chugoku :=
		[]string{
			"31",
			"32",
			"33",
			"34",
			"35",
		}

	shikoku :=
		[]string{
			"36",
			"37",
			"38",
			"39",
		}

	kyushu :=
		[]string{
			"40",
			"41",
			"42",
			"43",
			"44",
			"45",
			"46",
		}

	return []MerchantMaster{

		// 全国型 10社 × 200ホテル

		{
			ID:              "merchant_001",
			Name:            "Japan Stay Network 01",
			HotelCount:      200,
			PrefectureCodes: all,
		},
		{
			ID:              "merchant_002",
			Name:            "Japan Stay Network 02",
			HotelCount:      200,
			PrefectureCodes: all,
		},
		{
			ID:              "merchant_003",
			Name:            "Japan Stay Network 03",
			HotelCount:      200,
			PrefectureCodes: all,
		},
		{
			ID:              "merchant_004",
			Name:            "Japan Stay Network 04",
			HotelCount:      200,
			PrefectureCodes: all,
		},
		{
			ID:              "merchant_005",
			Name:            "Japan Stay Network 05",
			HotelCount:      200,
			PrefectureCodes: all,
		},
		{
			ID:              "merchant_006",
			Name:            "Japan Stay Network 06",
			HotelCount:      200,
			PrefectureCodes: all,
		},
		{
			ID:              "merchant_007",
			Name:            "Japan Stay Network 07",
			HotelCount:      200,
			PrefectureCodes: all,
		},
		{
			ID:              "merchant_008",
			Name:            "Japan Stay Network 08",
			HotelCount:      200,
			PrefectureCodes: all,
		},
		{
			ID:              "merchant_009",
			Name:            "Japan Stay Network 09",
			HotelCount:      200,
			PrefectureCodes: all,
		},
		{
			ID:              "merchant_010",
			Name:            "Japan Stay Network 10",
			HotelCount:      200,
			PrefectureCodes: all,
		},

		// 北海道

		{
			ID:              "merchant_011",
			Name:            "Hokkaido Stay 01",
			HotelCount:      20,
			PrefectureCodes: hokkaido,
		},
		{
			ID:              "merchant_012",
			Name:            "Hokkaido Stay 02",
			HotelCount:      20,
			PrefectureCodes: hokkaido,
		},

		// 沖縄

		{
			ID:              "merchant_013",
			Name:            "Okinawa Resort 01",
			HotelCount:      20,
			PrefectureCodes: okinawa,
		},
		{
			ID:              "merchant_014",
			Name:            "Okinawa Resort 02",
			HotelCount:      20,
			PrefectureCodes: okinawa,
		},

		// 東北

		{
			ID:              "merchant_015",
			Name:            "Tohoku Stay 01",
			HotelCount:      20,
			PrefectureCodes: tohoku,
		},
		{
			ID:              "merchant_016",
			Name:            "Tohoku Stay 02",
			HotelCount:      20,
			PrefectureCodes: tohoku,
		},

		// 関東

		{
			ID:              "merchant_017",
			Name:            "Kanto Stay 01",
			HotelCount:      20,
			PrefectureCodes: kanto,
		},
		{
			ID:              "merchant_018",
			Name:            "Kanto Stay 02",
			HotelCount:      20,
			PrefectureCodes: kanto,
		},
		{
			ID:              "merchant_019",
			Name:            "Kanto Stay 03",
			HotelCount:      20,
			PrefectureCodes: kanto,
		},

		// 東海

		{
			ID:              "merchant_020",
			Name:            "Tokai Stay 01",
			HotelCount:      20,
			PrefectureCodes: tokai,
		},
		{
			ID:              "merchant_021",
			Name:            "Tokai Stay 02",
			HotelCount:      20,
			PrefectureCodes: tokai,
		},

		// 近畿

		{
			ID:              "merchant_022",
			Name:            "Kinki Stay 01",
			HotelCount:      20,
			PrefectureCodes: kinki,
		},
		{
			ID:              "merchant_023",
			Name:            "Kinki Stay 02",
			HotelCount:      20,
			PrefectureCodes: kinki,
		},
		{
			ID:              "merchant_024",
			Name:            "Kinki Stay 03",
			HotelCount:      20,
			PrefectureCodes: kinki,
		},

		// 中国

		{
			ID:              "merchant_025",
			Name:            "Chugoku Stay 01",
			HotelCount:      20,
			PrefectureCodes: chugoku,
		},
		{
			ID:              "merchant_026",
			Name:            "Chugoku Stay 02",
			HotelCount:      20,
			PrefectureCodes: chugoku,
		},

		// 四国

		{
			ID:              "merchant_027",
			Name:            "Shikoku Stay 01",
			HotelCount:      20,
			PrefectureCodes: shikoku,
		},
		{
			ID:              "merchant_028",
			Name:            "Shikoku Stay 02",
			HotelCount:      20,
			PrefectureCodes: shikoku,
		},

		// 九州

		{
			ID:              "merchant_029",
			Name:            "Kyushu Stay 01",
			HotelCount:      20,
			PrefectureCodes: kyushu,
		},
		{
			ID:              "merchant_030",
			Name:            "Kyushu Stay 02",
			HotelCount:      20,
			PrefectureCodes: kyushu,
		},
	}
}

func AllPrefectureCodes() []string {

	result :=
		make(
			[]string,
			0,
			47,
		)

	for i := 1; i <= 47; i++ {

		if i < 10 {
			result = append(
				result,
				"0"+string(rune('0'+i)),
			)

			continue
		}

		result = append(
			result,
			twoDigit(i),
		)
	}

	return result
}

func twoDigit(
	value int,
) string {

	tens := value / 10
	ones := value % 10

	return string(
		[]byte{
			byte('0' + tens),
			byte('0' + ones),
		},
	)
}
