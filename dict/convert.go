package dict

import (
	"bytes"
	"index/suffixarray"
	"os"
)

// Convert ...
func Convert(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	defer file.Close()

	d := createDictReader(createTokenReader(file))

	var buffer bytes.Buffer
	_, _ = buffer.WriteRune('@')
	for !d.isDone() {
		g, err := d.nextGroup()
		if err != nil {
			return err
		}
		for _, w := range g.words {
			_, _ = buffer.WriteString(w.name)
			_, _ = buffer.WriteRune('@')
		}
	}
	sa := suffixarray.New(buffer.Bytes()[:])

	flags := os.O_CREATE | os.O_WRONLY
	dictFile, err := os.OpenFile("morph.dict", flags, 0666)
	if err != nil {
		return err
	}

	defer dictFile.Close()

	err = sa.Write(dictFile)
	if err != nil {
		return err
	}

	return nil

}
