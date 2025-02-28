package deck

import "strconv"

const _Suit_name = "SpadeDiamondClubHeartJoker"

var _Suit_index = [...]uint8{0, 5, 12, 16, 21, 26}

func (i Suit) String() string {
	if i >= Suit(len(_Suit_index)-1) {
		return "Suit(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Suit_name[_Suit_index[i]:_Suit_index[i+1]]
}

const _Rank_Name = "AceTwoThreeFourFiveSixSevenEightNineTenJackQueenKing"

var _Rank_index = [...]uint8{0, 3, 6 ,11, 15 ,19, 22, 27, 32, 36, 39, 43, 48, 52}

func (i Rank) String() string {
	i -= 1
	if i >= Rank(len(_Rank_index)-1) {
		return "Rank(" + strconv.FormatInt(int64(i+1), 10) +")"
	}
	return _Rank_Name[_Rank_index[i]:_Rank_index[i+1]]
}