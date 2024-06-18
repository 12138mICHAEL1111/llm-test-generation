package util

import (
	"encoding/gob"
	"os"
	openai "github.com/sashabaranov/go-openai"
)

func SaveSliceToFile(slice [][]openai.ChatCompletionMessage, filename string) error {
	if _, err := os.Stat(filename); err == nil {
		err := os.Remove(filename)
		if err != nil {
			panic(err)
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(slice); err != nil {
		return err
	}

	return nil
}

func LoadSliceFromFile(filename string) ([][]openai.ChatCompletionMessage, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var slice [][]openai.ChatCompletionMessage
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&slice); err != nil {
		return nil, err
	}

	return slice, nil
}
