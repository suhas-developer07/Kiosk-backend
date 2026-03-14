package utils

var TotalPages = 320

func UpdateRemainingPagesInTray(numberPagestodeduct int)(bool){
	TotalPages = TotalPages - numberPagestodeduct
	return true
}