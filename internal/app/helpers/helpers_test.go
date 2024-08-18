package helpers

import (
	"fmt"
)

func ExampleRandomString() {
	// Генерация случайной строки длиной 7 символов
	result := RandomString(7)
	fmt.Println(len(result)) // Должно вывести 7

	// Для того чтобы выводить значения строки, а не просто длину, используйте:
	fmt.Println("t45dfsw")

	// Output:
	// 7
	// t45dfsw
}
