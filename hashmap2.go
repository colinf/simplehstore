package simplehstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// HashMap2 contains a KeyValue struct and a dbDatastructure.
// Each value is a JSON data blob and can contains sub-keys.
type HashMap2 struct {
	dbDatastructure        // KeyValue is .host *Host + .table string
	seenPropTable   string // Set of all encountered property keys
}

// A string that is unlikely to appear in a key
const fieldSep = "¤"

// NewHashMap2 creates a new HashMap2 struct
func NewHashMap2(host *Host, name string) (*HashMap2, error) {
	var hm2 HashMap2
	// kv is a KeyValue (HSTORE) table of all properties (key = owner_ID + "¤" + property_key)
	kv, err := NewKeyValue(host, name+"_properties_HSTORE_map")
	if err != nil {
		return nil, err
	}
	// seenPropSet is a set of all encountered property keys
	seenPropSet, err := NewSet(host, name+"_encountered_property_keys")
	if err != nil {
		return nil, err
	}
	hm2.host = host
	hm2.table = kv.table
	hm2.seenPropTable = seenPropSet.table
	return &hm2, nil
}

// KeyValue returns the *KeyValue of properties for this HashMap2
func (hm2 *HashMap2) KeyValue() *KeyValue {
	return &KeyValue{hm2.host, hm2.table}
}

// PropSet returns the property *Set for this HashMap2
func (hm2 *HashMap2) PropSet() *Set {
	return &Set{hm2.host, hm2.seenPropTable}
}

// Set a value in a hashmap given the element id (for instance a user id) and the key (for instance "password")
func (hm2 *HashMap2) Set(owner, key, value string) error {
	return hm2.SetMap(owner, map[string]string{key: value})
}

// setPropWithTransaction will set a value in a hashmap given the element id (for instance a user id) and the key (for instance "password")
func (hm2 *HashMap2) setPropWithTransaction(ctx context.Context, transaction *sql.Tx, owner, key, value string, checkForFieldSep bool) error {
	if checkForFieldSep {
		if strings.Contains(owner, fieldSep) {
			return fmt.Errorf("owner can not contain %s", fieldSep)
		}
		if strings.Contains(key, fieldSep) {
			return fmt.Errorf("key can not contain %s", fieldSep)
		}
	}
	// Add the key to the property set, without using a transaction
	if err := hm2.PropSet().Add(key); err != nil {
		return err
	}
	// Set a key + value for this "owner¤key"
	kv := hm2.KeyValue()
	if !kv.host.rawUTF8 {
		Encode(&value)
	}
	encodedValue := value
	return kv.setWithTransaction(ctx, transaction, owner+fieldSep+key, encodedValue)
}

// SetMap will set many keys/values, in a single transaction
func (hm2 *HashMap2) SetMap(owner string, m map[string]string) error {
	checkForFieldSep := true

	// Get all properties
	propset := hm2.PropSet()
	allProperties, err := propset.All()
	if err != nil {
		return err
	}

	// Use a context and a transaction to bundle queries
	ctx := context.Background()
	transaction, err := hm2.host.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Prepare the changes
	for k, v := range m {
		if err := hm2.setPropWithTransaction(ctx, transaction, owner, k, v, checkForFieldSep); err != nil {
			transaction.Rollback()
			return err
		}
		if !hasS(allProperties, k) {
			if err := propset.Add(k); err != nil {
				transaction.Rollback()
				return err
			}
		}
	}
	return transaction.Commit()
}

// SetLargeMap will add many owners+keys/values, in a single transaction, without checking if they already exists.
// It also does not check if the keys or property keys contains fieldSep (¤) or not, for performance.
// These must all be brand new "usernames" (the first key), and not be in the existing hm2.OwnerSet().
// This function has good performance, but must be used carefully.
func (hm2 *HashMap2) SetLargeMap(allProperties map[string]map[string]string) error {

	// First get the KeyValue and Set structures that will be used
	kv := hm2.KeyValue()
	propSet := hm2.PropSet()

	// All seen properties
	props, err := propSet.All()
	if err != nil {
		return err
	}

	// Find new properties in the allProperties map
	var newProps []string
	for owner := range allProperties {
		// Find all unique properties
		for k := range allProperties[owner] {
			if !hasS(props, k) && !hasS(newProps, k) {
				newProps = append(newProps, k)
			}
		}
	}

	// Store the new properties
	for _, prop := range newProps {
		if Verbose {
			fmt.Printf("ADDING %s\n", prop)
		}
		if err := propSet.Add(prop); err != nil {
			return err
		}
	}

	ctx := context.Background()

	// Start one goroutine + transaction for the recognized owners

	if Verbose {
		fmt.Println("Starting transaction")
	}

	// Create a new transaction
	transaction, err := hm2.host.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Then update all recognized owners
	for owner, propMap := range allProperties {
		// Prepare the changes
		for k, value := range propMap {
			if Verbose {
				fmt.Printf("SETTING %s %s->%s\n", owner, k, value)
			}
			// Set a key + value for this "owner¤key"
			if !kv.host.rawUTF8 {
				Encode(&value)
			}
			encodedValue := value
			if _, err := kv.updateWithTransaction(ctx, transaction, owner+fieldSep+k, encodedValue); err != nil {
				transaction.Rollback()
				return err
			}
		}
	}

	if Verbose {
		fmt.Println("Committing transaction")
	}
	if err := transaction.Commit(); err != nil {
		return err
	}

	fmt.Println("Transaction complete")

	return nil // success
}

// Get a value.
// Returns: value, error
// If a value was not found, an empty string is returned.
func (hm2 *HashMap2) Get(owner, key string) (string, error) {
	return hm2.KeyValue().Get(owner + fieldSep + key)
}

// Get multiple values
func (hm2 *HashMap2) GetMap(owner string, keys []string) (map[string]string, error) {
	results := make(map[string]string)

	// Use a context and a transaction to bundle queries
	ctx := context.Background()
	transaction, err := hm2.host.db.BeginTx(ctx, nil)
	if err != nil {
		return results, err
	}

	for _, key := range keys {
		s, err := hm2.KeyValue().getWithTransaction(ctx, transaction, owner+fieldSep+key)
		if err != nil {
			transaction.Rollback()
			return results, err
		}
		results[key] = s
	}

	transaction.Commit()
	return results, nil
}

// Has checks if a given owner + key exists in the hash map
func (hm2 *HashMap2) Has(owner, key string) (bool, error) {
	s, err := hm2.KeyValue().Get(owner + fieldSep + key)
	if err != nil {
		if noResult(err) {
			// Not an actual error, just got no results
			return false, nil
		}
		// An actual error
		return false, err
	}
	// No error, got a result
	if s == "" {
		return false, nil
	}
	return true, nil
}

// Exists checks if a given owner exists as a hash map at all.
func (hm2 *HashMap2) Exists(owner string) (bool, error) {
	// Looking up the owner directly is tricky, but with a property, it's faster.
	allProps, err := hm2.PropSet().All()
	if err != nil {
		return false, err
	}
	// TODO: Improve the performance of this by using SQL instead of looping
	for _, key := range allProps {
		if found, err := hm2.Has(owner, key); err == nil && found {
			return true, nil
		}
	}
	return false, nil
}

// AllWhere returns all owner ID's that has a property where key == value
func (hm2 *HashMap2) AllWhere(key, value string) ([]string, error) {
	allOwners, err := hm2.All()
	if err != nil {
		return []string{}, err
	}
	// TODO: Improve the performance of this by using SQL instead of looping
	foundOwners := []string{}
	for _, owner := range allOwners {
		// The owner+key exists and the value matches the given value
		if v, err := hm2.Get(owner, key); err == nil && v == value {
			foundOwners = append(foundOwners, owner)
		}
	}
	return foundOwners, nil
}

// AllEncounteredKeys returns all encountered keys for all owners
func (hm2 *HashMap2) AllEncounteredKeys() ([]string, error) {
	return hm2.PropSet().All()
}

// Keys loops through absolutely all owners and all properties in the database
// and returns all found keys.
func (hm2 *HashMap2) Keys(owner string) ([]string, error) {
	allProps, err := hm2.PropSet().All()
	if err != nil {
		return []string{}, err
	}
	// TODO: Improve the performance of this by using SQL instead of looping
	allKeys := []string{}
	for _, key := range allProps {
		if found, err := hm2.Has(owner, key); err == nil && found {
			allKeys = append(allKeys, key)
		}
	}
	return allKeys, nil
}

// All returns all owner ID's
func (hm2 *HashMap2) All() ([]string, error) {
	foundOwners := make(map[string]bool)
	allOwnersAndKeys, err := hm2.KeyValue().All()
	if err != nil {
		return []string{}, err
	}
	for _, ownerAndKey := range allOwnersAndKeys {
		if pos := strings.Index(ownerAndKey, fieldSep); pos != -1 {
			owner := ownerAndKey[:pos]
			if _, has := foundOwners[owner]; !has {
				foundOwners[owner] = true
			}
		}
	}
	keys := make([]string, len(foundOwners))
	i := 0
	for k := range foundOwners {
		keys[i] = k
		i++
	}
	return keys, nil
}

// Count counts the number of owners for hash map elements
func (hm2 *HashMap2) Count() (int, error) {
	a, err := hm2.All()
	if err != nil {
		return 0, err
	}
	return len(a), nil
	// return hm2.KeyValue().Count() is not correct, since it counts all owners + fieldSep + keys

}

// CountInt64 counts the number of owners for hash map elements (int64)
func (hm2 *HashMap2) CountInt64() (int64, error) {
	a, err := hm2.All()
	if err != nil {
		return 0, err
	}
	return int64(len(a)), nil
	// return hm2.KeyValue().Count() is not correct, since it counts all owners + fieldSep + keys
}

// DelKey removes a key of an owner in a hashmap (for instance the email field for a user)
func (hm2 *HashMap2) DelKey(owner, key string) error {
	// The key is not removed from the set of all encountered properties
	// even if it's the last key with that name, for a performance vs storage tradeoff.
	return hm2.KeyValue().Del(owner + fieldSep + key)
}

// Del removes an element (for instance a user)
func (hm2 *HashMap2) Del(owner string) error {
	allProps, err := hm2.PropSet().All()
	if err != nil {
		return err
	}
	for _, key := range allProps {
		if err := hm2.KeyValue().Del(owner + fieldSep + key); err != nil {
			return err
		}
	}
	return nil
}

// Remove this hashmap
func (hm2 *HashMap2) Remove() error {
	hm2.PropSet().Remove()
	if err := hm2.KeyValue().Remove(); err != nil {
		return fmt.Errorf("could not remove kv: %s", err)
	}
	return nil
}

// Clear the contents
func (hm2 *HashMap2) Clear() error {
	hm2.PropSet().Clear()
	if err := hm2.KeyValue().Clear(); err != nil {
		return err
	}
	return nil
}
