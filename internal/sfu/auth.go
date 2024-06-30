package sfu

var (
	storage = map[string]*User{
		"OqTftBr8MUKXwqK2eTLysGHP=4jpDw0glOGcq8=WqH8H8GhklrLKvZT4XTTA7LqM": {
			Name:     "Daniil",
			Lastname: "Naumovets",
		},
		"K-W?8zBePKhWePJZf081bx/6fljB1msCGe9N8owOCIyE!zS7N1XU2hybI9VCox4G": {
			Name:     "Kirill",
			Lastname: "Prohorov",
		},
		"OF59xe8D3!Qw-CyHklhVaZ2DotFNX=n5aONKSOyDZbcZFoLti-OWp?Pk8Y4!R7/u": {
			Name:     "John",
			Lastname: "Soliev",
		},
	}
)

func Auth(token string) (*User, bool) {
	user, ok := storage[token]

	return user, ok
}
