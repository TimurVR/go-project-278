package dto


//от меня
type LinkResponce struct{
	Id 				int		`json:"id"`
	Original_url 	string	`json:"original_url"`
	Short_name 		string	`json:"short_name"`
	Short_url 		string	`json:"short_url"`
}
