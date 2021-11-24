package simplehstore

import (
	"fmt"
	"testing"

	// For testing the storage of bcrypt password hashes
	"golang.org/x/crypto/bcrypt"

	"crypto/sha256"
	"io"

	"github.com/xyproto/cookie"
	"github.com/xyproto/pinterface"
)

func TestHashMap2UserStateShort(t *testing.T) {
	Verbose = true

	//host := New() // locally
	host := NewHost(defaultConnectionString)

	defer host.Close()
	hashmap, err := NewHashMap2(host, hashmapname)
	if err != nil {
		t.Error(err)
	}

	hashmap.Clear()

	username := "bob"

	err = hashmap.Set(username, "aa", "true")
	if err != nil {
		t.Error(err)
	}

	err = hashmap.Set(username, "aa", "false")
	if err != nil {
		t.Error(err)
	}

	err = hashmap.Set(username, "bb", "82")
	if err != nil {
		t.Error(err)
	}

	aval, err := hashmap.Get(username, "aa")
	if err != nil {
		t.Error(err)
	}
	if aval != "false" {
		t.Error("aa should be false, but it is: " + aval)
	}

	err = hashmap.SetMap(username, map[string]string{"x": "42", "y": "64"})
	if err != nil {
		t.Error(err)
	}

	aval, err = hashmap.Get(username, "x")
	if err != nil {
		t.Error(err)
	}
	if aval != "42" {
		t.Errorf("expected 42, got %s", aval)
	}

	aval, err = hashmap.Get(username, "y")
	if err != nil {
		t.Error(err)
	}
	if aval != "64" {
		t.Errorf("expected 64, got %s", aval)
	}

	err = hashmap.Remove()
	if err != nil {
		t.Errorf("Error, could not remove hashmap! %s", err.Error())
	}
}

func TestHashMap2UserState(t *testing.T) {
	Verbose = true

	//host := New() // locally
	host := NewHost(defaultConnectionString)

	defer host.Close()

	hashmap, err := NewHashMap2(host, hashmapname)
	if err != nil {
		t.Error(err)
	}
	hashmap.Clear()

	username := "bob"

	err = hashmap.Set(username, "a", "true")
	if err != nil {
		t.Error(err)
	}

	err = hashmap.Set(username, "a", "false")
	if err != nil {
		t.Error(err)
	}

	aval, err := hashmap.Get(username, "a")
	if err != nil {
		t.Error(err)
	}
	if aval != "false" {
		t.Error("a should be false")
	}

	err = hashmap.Set(username, "a", "true")
	if err != nil {
		t.Error(err)
	}

	err = hashmap.Set(username, "b", "true")
	if err != nil {
		t.Error(err)
	}

	err = hashmap.Set(username, "b", "true")
	if err != nil {
		t.Error(err)
	}

	aval, err = hashmap.Get(username, "a")
	if err != nil {
		t.Errorf("Error when retrieving element! %s", err.Error())
	}
	if aval != "true" {
		t.Error("a should be true")
	}

	bval, err := hashmap.Get(username, "b")
	if err != nil {
		t.Errorf("Error when retrieving elements! %s", err.Error())
	}
	if bval != "true" {
		t.Error("b should be true")
	}

	err = hashmap.Remove()
	if err != nil {
		t.Errorf("Error, could not remove hashmap! %s", err.Error())
	}

}

func TestHash2KvMix(t *testing.T) {
	Verbose = true

	//host := New() // locally
	host := NewHost(defaultConnectionString)

	defer host.Close()

	sameName := "ostekake"

	h, err := NewHashMap2(host, sameName)
	if err != nil {
		t.Error(err)
	}
	h.Set("a", "b", "c")
	defer h.Remove()

	kv, err := NewKeyValue(host, sameName)
	if err != nil {
		t.Error(err)
	}
	kv.Remove()

	v, err := h.Get("a", "b")
	if err != nil {
		t.Error(err)
	}

	if v != "c" {
		t.Errorf("Error, hashmap table name collision")
	}
}

func TestHash2Storage(t *testing.T) {
	Verbose = true

	//host := New() // locally
	host := NewHost(defaultConnectionString)

	defer host.Close()
	hashmap, err := NewHashMap2(host, hashmapname)
	if err != nil {
		t.Error(err)
	}
	hashmap.Clear()

	username := "bob"
	key := "password"
	password := "hunter1"

	// bcrypt test

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	value := string(passwordHash)

	if err := hashmap.Set(username, key, value); err != nil {
		t.Errorf("Error, could not set value in hashmap! %s", err.Error())
	}
	item, err := hashmap.Get(username, key)
	if err != nil {
		t.Errorf("Unable to retrieve from hashmap! %s\n", err.Error())
	}
	if item != value {
		t.Errorf("Error, got a different value back (bcrypt)! %s != %s\n", value, item)
	}

	// sha256 test

	hasher := sha256.New()
	io.WriteString(hasher, password+cookie.RandomCookieFriendlyString(30)+username)
	passwordHash = hasher.Sum(nil)
	value = string(passwordHash)

	if err := hashmap.Set(username, key, value); err != nil {
		t.Errorf("Error, could not set value in hashmap! %s", err.Error())
	}
	item, err = hashmap.Get(username, key)
	if err != nil {
		t.Errorf("Unable to retrieve from hashmap! %s\n", err.Error())
	}
	if item != value {
		t.Errorf("Error, got a different value back (sha256)! %s != %s\n", value, item)
	}

	err = hashmap.Remove()
	if err != nil {
		t.Errorf("Error, could not remove hashmap! %s", err.Error())
	}
}

// Check that "bob" is confirmed
func TestConfirmed2(t *testing.T) {
	host := NewHost(defaultConnectionString)
	defer host.Close()
	users, err := NewHashMap2(host, "users")
	if err != nil {
		t.Error(err)
	}
	defer users.Remove()
	users.Set("bob", "confirmed", "true")
	ok, err := users.Exists("bob")
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("bob should exist!")
	}
	val, err := users.Get("bob", "confirmed")
	if err != nil {
		t.Error(err)
	}
	if val != "true" {
		t.Error("bob should be confirmed")
	}
	err = users.DelKey("bob", "confirmed")
	if err != nil {
		t.Error(err)
	}
	ok, err = users.Has("bob", "confirmed")
	if err != nil {
		t.Error(err)
	}
	if ok {
		t.Error("The confirmed key should be gone")
	}
}

func TestHashMap2UserState2(t *testing.T) {
	Verbose = true

	//host := New() // locally
	host := NewHost(defaultConnectionString)

	defer host.Close()
	hashmap, err := NewHashMap2(host, hashmapname)
	if err != nil {
		t.Error(err)
	}
	hashmap.Clear()

	username := "bob"
	key := "password"
	value := "hunter1"

	// Get key that doesn't exist yet
	_, err = hashmap.Get("ownerblabla", "keyblabla")
	if err == nil {
		t.Errorf("Key found, when it should be missing! %s", err.Error())
	}

	if err := hashmap.Set(username, key, value); err != nil {
		t.Errorf("Error, could not set value in hashmap! %s", err.Error())
	}

	hashmap.Remove()
}

func TestHashMap2(t *testing.T) {
	Verbose = true

	//host := New() // locally
	host := NewHost(defaultConnectionString)

	defer host.Close()
	hashmap, err := NewHashMap2(host, hashmapname)
	if err != nil {
		t.Error(err)
	}
	hashmap.Clear()

	username := "bob"
	key := "password"
	value := "hunter1"

	// Get key that doesn't exist yet
	_, err = hashmap.Get("ownerblabla", "keyblabla")
	if err == nil {
		t.Errorf("Key found, when it should be missing! %s", err.Error())
	}

	if err := hashmap.Set(username, key, value); err != nil {
		t.Errorf("Error, could not set value in hashmap! %s", err.Error())
	}

	// Once more, with the same data
	if err := hashmap.Set(username, key, value); err != nil {
		t.Errorf("Error, could not set value in hashmap! %s", err.Error())
	}

	items, err := hashmap.All()
	if err != nil {
		t.Errorf("Error when retrieving elements! %s", err.Error())
	}
	if len(items) != 1 {
		t.Errorf("Error, wrong element length! %v", len(items))
	}

	// Add one more item, so that there are 2 entries in the database
	if err := hashmap.Set("bob", "number", "42"); err != nil {
		t.Errorf("Error, could not set value in hashmap! %s", err.Error())
	}

	// Add one more item, so that there are 3 entries in the database,
	// two with owner "bob" and 1 with owner "alice"
	if err := hashmap.Set("alice", "number", "42"); err != nil {
		t.Errorf("Error, could not set value in hashmap! %s", err.Error())
	}

	// Retrieve items again and check the length
	items, err = hashmap.All()
	if err != nil {
		t.Errorf("Error, could not retrieve all items! %s", err.Error())
	}
	if len(items) != 2 {
		for i, item := range items {
			fmt.Printf("ITEM %d IS %v\n", i, item)
		}
		t.Errorf("Error, wrong element length! %v", len(items))
	}

	//if (len(items) > 0) && (items[0] != username) {
	//	t.Errorf("Error, wrong elementid! %v", items)
	//}

	item, err := hashmap.Get(username, key)
	if err != nil {
		t.Errorf("Error, could not fetch value from hashmap! %s", err.Error())
	}
	if item != value {
		t.Errorf("Error, expected %s, got %s!", value, item)
	}

	count, err := hashmap.Count()
	if err != nil {
		t.Error("Error, could not get the count!")
	}
	if count != 2 {
		t.Errorf("Error, expected the count of bob and alice to be 2, got %d!", count)
	}

	items, err = hashmap.AllWhere("number", "42")
	if err != nil {
		t.Error("Error, could not get value for property number")
	}
	if len(items) == 0 {
		t.Error("Error, there should be more than 0 entries for the number property")
	}
	fmt.Println("Items where number is 42:", items)

	// Delete the "number" property/key from owner "bob"
	err = hashmap.DelKey("bob", "number")
	if err != nil {
		t.Error(err)
	}

	// Delete the "number" property/key from owner "alice"
	err = hashmap.DelKey("alice", "number")
	if err != nil {
		t.Error(err)
	}

	keys, err := hashmap.Keys(username)
	if err != nil {
		t.Error(err)
	}

	if len(keys) == 0 {
		t.Errorf("Error, keys for %s are empty but should contain %s\n", username, "password")
	}

	// only "password"
	if len(keys) != 1 {
		t.Errorf("Error, wrong keys: %v\n", keys)
	}
	if keys[0] != "password" {
		t.Errorf("Error, wrong keys: %v\n", keys)
	}

	err = hashmap.Remove()
	if err != nil {
		t.Errorf("Error, could not remove hashmap! %s", err.Error())
	}

	// Check that hashmap qualifies for the IHashMap interface
	var _ pinterface.IHashMap = hashmap
}

func TestDashesAndQuotes2(t *testing.T) {
	Verbose = true

	//host := New() // locally
	host := NewHost(defaultConnectionString)

	defer host.Close()
	hashmap, err := NewHashMap2(host, hashmapname+"'s-")
	if err != nil {
		t.Error(err)
	}
	hashmap.Clear()

	username := "bob's kitchen-machine"
	key := "password"
	value := "hunter's table-cloth"

	// Get key that doesn't exist yet
	_, err = hashmap.Get("ownerblabla", "keyblabla")
	if err == nil {
		t.Errorf("Key found, when it should be missing! %s", err.Error())
	}

	if err := hashmap.Set(username, key, value); err != nil {
		t.Errorf("Error, could not set value in hashmap! %s", err.Error())
	}
	// Once more, with the same data
	if err := hashmap.Set(username, key, value); err != nil {
		t.Errorf("Error, could not set value in hashmap! %s", err.Error())
	}
	if _, err := hashmap.All(); err != nil {
		t.Errorf("Error when retrieving elements! %s", err.Error())
	}
	item, err := hashmap.Get(username, key)
	if err != nil {
		t.Errorf("Error, could not fetch value from hashmap! %s", err.Error())
	}
	if item != value {
		t.Errorf("Error, expected %s, got %s!", value, item)
	}
	err = hashmap.Remove()
	if err != nil {
		t.Errorf("Error, could not remove hashmap! %s", err.Error())
	}
}
