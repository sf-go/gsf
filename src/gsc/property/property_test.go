package property

import (
	"gsc/serialization"
	"testing"
)

type user struct {
	*Property

	Name string
	Age  int
	Dot  float64
}

func newUser() *user {
	return &user{
		Property: NewProperty(),
	}
}

func TestValueType(t *testing.T) {
	user := newUser()
	user.Register(&user)

	user.SetValue("Name", "sss")
	t.Log(user.GetValue("Name"))

	user.SetValue("Age", 100)
	t.Log(user.GetValue("Age"))

	user.SetValue("Dot", 3.1415)
	t.Log(user.GetValue("Dot"))
}

func TestSerialization(t *testing.T) {
	user := newUser()
	user.Register(&user)

	user.Name = "xxx"
	user.Dot = 3.1415

	user.SetValue("Name", "sss")
	t.Log(user.GetValue("Name"))

	user.SetValue("Age", 100)
	t.Log(user.GetValue("Age"))

	write := serialization.NewEndianBinaryWriter()
	user.ToBinaryWriter(write)
	bytes := write.ToBytes()

	user.Name = ""
	user.Age = 0

	reader := serialization.NewEndianBinaryReader(bytes)
	user.FromBinaryReader(reader)

	t.Log(user.Name)
	t.Log(user.Age)
	t.Log(user.Dot)
}