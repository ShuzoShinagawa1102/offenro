package mockdata

import (
	"fmt"
	"hash/fnv"
	"time"
)

type Prefecture struct {
	Code string
	Name string
	City string
}

type Hotel struct {
	HotelID        string
	MerchantID     string
	Name           string
	PrefectureCode string
	PrefectureName string
	City           string
}

type Store struct {
	merchants map[string]MerchantMaster

	hotelsByMerchant map[string][]Hotel
}

func NewStore() *Store {

	store :=
		&Store{
			merchants:
				make(
					map[string]MerchantMaster,
				),

			hotelsByMerchant:
				make(
					map[string][]Hotel,
				),
		}

	for _, merchant :=
		range MerchantMasters() {

		store.merchants[merchant.ID] = merchant

		store.hotelsByMerchant[merchant.ID] = generateHotels(
			merchant,
		)
	}

	return store
}

func (s *Store) HasMerchant(
	merchantID string,
) bool {

	_, exists :=
		s.merchants[merchantID]

	return exists
}

func (s *Store) Hotels(
	merchantID string,
) []Hotel {

	return s.hotelsByMerchant[merchantID]
}

func generateHotels(
	merchant MerchantMaster,
) []Hotel {

	hotels :=
		make(
			[]Hotel,
			0,
			merchant.HotelCount,
		)

	for i := 0;
		i < merchant.HotelCount;
		i++ {

		prefectureCode :=
			merchant.
				PrefectureCodes[
				i%
					len(
						merchant.
							PrefectureCodes,
					),
			]

		prefecture, exists :=
			PrefectureByCode(
				prefectureCode,
			)

		if !exists {
			continue
		}

		number := i + 1

		hotels = append(
			hotels,
			Hotel{
				HotelID:
					fmt.Sprintf(
						"%s-h%03d",
						merchant.ID,
						number,
					),

				MerchantID:
					merchant.ID,

				Name:
					fmt.Sprintf(
						"%s %s Hotel %03d",
						merchant.Name,
						prefecture.Name,
						number,
					),

				PrefectureCode:
					prefecture.Code,

				PrefectureName:
					prefecture.Name,

				City:
					prefecture.City,
			},
		)
	}

	return hotels
}

// AvailabilityForはMock用。
// hotelID + dateから決定論的に
// 空室数と1室1泊価格を算出する。
//
// 実Merchantでは当然、自社DB・予約基盤などから取得する。
func AvailabilityFor(
	hotelID string,
	date time.Time,
) (
	availableRooms int,
	pricePerRoom int64,
) {

	hash :=
		fnv.New32a()

	_, _ =
		hash.Write(
			[]byte(
				hotelID +
					"|" +
					date.Format(
						"2006-01-02",
					),
			),
		)

	value := hash.Sum32()

	availableRooms =
		int(
			value % 6,
		)

	pricePerRoom =
		8000 +
			int64(
				(value/7)%22000,
			)

	return
}

func PrefectureByCode(
	code string,
) (
	Prefecture,
	bool,
) {

	prefecture, exists :=
		prefectures[code]

	return prefecture, exists
}

var prefectures =
	map[string]Prefecture{

		"01": {"01", "北海道", "札幌市"},
		"02": {"02", "青森県", "青森市"},
		"03": {"03", "岩手県", "盛岡市"},
		"04": {"04", "宮城県", "仙台市"},
		"05": {"05", "秋田県", "秋田市"},
		"06": {"06", "山形県", "山形市"},
		"07": {"07", "福島県", "福島市"},

		"08": {"08", "茨城県", "水戸市"},
		"09": {"09", "栃木県", "宇都宮市"},
		"10": {"10", "群馬県", "前橋市"},
		"11": {"11", "埼玉県", "さいたま市"},
		"12": {"12", "千葉県", "千葉市"},
		"13": {"13", "東京都", "新宿区"},
		"14": {"14", "神奈川県", "横浜市"},

		"15": {"15", "新潟県", "新潟市"},
		"16": {"16", "富山県", "富山市"},
		"17": {"17", "石川県", "金沢市"},
		"18": {"18", "福井県", "福井市"},
		"19": {"19", "山梨県", "甲府市"},
		"20": {"20", "長野県", "長野市"},

		"21": {"21", "岐阜県", "岐阜市"},
		"22": {"22", "静岡県", "静岡市"},
		"23": {"23", "愛知県", "名古屋市"},
		"24": {"24", "三重県", "津市"},

		"25": {"25", "滋賀県", "大津市"},
		"26": {"26", "京都府", "京都市"},
		"27": {"27", "大阪府", "大阪市"},
		"28": {"28", "兵庫県", "神戸市"},
		"29": {"29", "奈良県", "奈良市"},
		"30": {"30", "和歌山県", "和歌山市"},

		"31": {"31", "鳥取県", "鳥取市"},
		"32": {"32", "島根県", "松江市"},
		"33": {"33", "岡山県", "岡山市"},
		"34": {"34", "広島県", "広島市"},
		"35": {"35", "山口県", "山口市"},

		"36": {"36", "徳島県", "徳島市"},
		"37": {"37", "香川県", "高松市"},
		"38": {"38", "愛媛県", "松山市"},
		"39": {"39", "高知県", "高知市"},

		"40": {"40", "福岡県", "福岡市"},
		"41": {"41", "佐賀県", "佐賀市"},
		"42": {"42", "長崎県", "長崎市"},
		"43": {"43", "熊本県", "熊本市"},
		"44": {"44", "大分県", "大分市"},
		"45": {"45", "宮崎県", "宮崎市"},
		"46": {"46", "鹿児島県", "鹿児島市"},
		"47": {"47", "沖縄県", "那覇市"},
	}
