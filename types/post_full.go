package types

// Полная информация о сообщении, включая связанные объекты.
// Тут используются указатели, что бы корректно отображало отсутствующие поля.
// при указателе не отрисовывается, при значении {"author": {}} добавляет лишний ключ.
//easyjson:json
type PostFull struct {
	Author *User   `json:"author,omitempty"`
	Forum  *Forum  `json:"forum,omitempty"`
	Post   *Post   `json:"post"`
	Thread *Thread `json:"thread,omitempty"`
}
