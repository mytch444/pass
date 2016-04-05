package main

import (
	"fmt"
	"os"
	"crypto/rand"
)

type Secstore struct {
	partRoot *Part
	Pwd *Part
}

func ParseSecstore(bytes []byte) (*Secstore, error) {
	var err error
	var store *Secstore

	store = new(Secstore)

	store.partRoot = new(Part)
	store.partRoot.Name = "/"
	
	store.partRoot.SubParts, _, err = ParseParts(bytes, store.partRoot)
	if err != nil {
		return nil, err
	}
	
	store.Pwd = store.partRoot

	return store, nil
}

func (store *Secstore) ToBytes() []byte {
	var bytes []byte = []byte(nil)

	for part := store.partRoot.SubParts; part != nil; part = part.Next {
		bytes = append(bytes, part.ToBytes()...)
	}

	return bytes
}

func (store *Secstore) FindPart(name string) *Part {
	return store.Pwd.FindSub(name)
}

func (store *Secstore) ChangeDir(name string) {
	var n *Part

	if name == ".." {
		n = store.Pwd.Parent
	} else {
		n = store.FindPart(name)
	}

	if n != nil {
		store.Pwd = n
	} else {
		fmt.Fprintln(os.Stderr, name, "does not exist")
	}
}

func (store *Secstore) RemovePart(name string) {
	var p, part, parent *Part
	var path string

	part = store.FindPart(name)
	if part == nil {
		fmt.Fprintln(os.Stderr, name, "does not exist")
		return
	}
	
	path, _ = splitLast(name, '/')
	if len(path) > 0 {
		parent = store.Pwd.FindSub(path)
	} else {
		parent = store.Pwd
	}

	fmt.Fprintln(os.Stderr, "Removing", name)

	if parent.SubParts == part {
		parent.SubParts = part.Next
	} else {
		for p = parent.SubParts; p != nil; p = p.Next {
			if p.Next == part {
				p.Next = part.Next
			}
		}
	}
}

func (store *Secstore) ShowPart(name string) {
	part := store.FindPart(name)
	if part == nil {
		fmt.Println(name, "not found")
	} else {
		part.Print()
	}
}

func (store *Secstore) List() {
	store.Pwd.Print()
}

func (store *Secstore) EditPart(name string) {
	var part *Part
	var err error
	var data string

	part = store.FindPart(name)
	if part == nil {
		fmt.Println(name, "not found.")
	} else if part.Data == "" {
		fmt.Println(name, "is a directory.")
	} else {
		data, err = OpenEditor(part.Data)
		if err != nil {
			fmt.Println("Not saving. Error running editor:", err)
		} else {
			part.Data = data
		}
	}
}

func splitLast(s string, sep rune) (main, last string) {
	for i := len(s) - 1; i >= 0; i-- {
		if rune(s[i]) == sep {
			return s[:i], s[i+1:]
		}
	}

	return "", s
}

func (store *Secstore) addPart(fpath string) (*Part, error) {
	var part, parent *Part
	var path, name string

	part = store.FindPart(fpath)
	if part != nil {
		return nil, fmt.Errorf("%s already exists", fpath)
	}
	
	path, name = splitLast(fpath, '/')
	if len(path) > 0 {
		parent = store.Pwd.FindSub(path)
	} else {
		parent = store.Pwd
	}

	part = new(Part)
	part.Name = name
	part.Parent = parent

	part.Next = parent.SubParts
	parent.SubParts = part

	return part, nil
}

func randomPass() string {
	var sum, r int
	var b []byte
	var err error

	b = make([]byte, 24)
	_, err = rand.Read(b)

	if err != nil {
		return "Error generating random bytes!"
	} 

	sum = 0
	for i := 0; i < len(b); i++ {
		sum += int(b[i])
		r = sum % 3
		switch (r) {
		case 0:
			b[i] = 'a' + b[i] % 26
		case 1:
			b[i] = 'A' + b[i] % 26
		case 2:
			b[i] = '0' + b[i] % 10
		}
	}

	return string(b)
}

func (store *Secstore) MakeNewPart(name string) {
	part, err := store.addPart(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	part.Data, _ = OpenEditor(randomPass())
	fmt.Println("Adding password:", name)
}

func (store *Secstore) MakeNewDirPart(name string) {
	part, err := store.addPart(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	part.Data = ""
	part.SubParts = nil

	fmt.Println("Adding directory:", name)
}
