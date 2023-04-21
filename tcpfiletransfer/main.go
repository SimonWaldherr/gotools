// Description: A simple TCP file transfer program.
package main

import (
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

// encrypt encrypts the input data using AES-GCM with the provided key.
func encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, data, nil)
	return append(nonce, ciphertext...), nil
}

// decrypt decrypts the input data using AES-GCM with the provided key.
func decrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(data) < 12 {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:12], data[12:]
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// compressFile compresses the input file using gzip.
func compressFile(input *os.File) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", "compressed_*.gz")
	if err != nil {
		return nil, err
	}

	gzipWriter := gzip.NewWriter(tmpFile)
	_, err = io.Copy(gzipWriter, input)
	if err != nil {
		return nil, err
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}

	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return tmpFile, nil
}

// decompressFile decompresses the input file using gzip.
func decompressFile(input *os.File) (*os.File, error) {
	tmpFile, err := os.CreateTemp("", "decompressed_*")
	if err != nil {
		return nil, err
	}

	gzipReader, err := gzip.NewReader(input)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(tmpFile, gzipReader)
	if err != nil {
		return nil, err
	}

	err = gzipReader.Close()
	if err != nil {
		return nil, err
	}

	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return tmpFile, nil
}

func main() {
	// Command line flags
	mode := flag.String("mode", "", "Server or client mode ('server' or 'client')")
	addr := flag.String("addr", "localhost:8080", "Address to listen on or connect to")
	filePath := flag.String("file", "", "Path to the file to send or receive")
	compress := flag.Bool("compress", false, "Enable compression")
	encryptFlag := flag.Bool("encrypt", false, "Enable encryption")
	key := flag.String("key", "", "Encryption key (required if encryption is enabled)")
	validate := flag.Bool("validate", false, "Enable validation (compute and send/receive SHA-256 hash)")

	flag.Parse()

	if *mode == "" || *filePath == "" {
		fmt.Println("Error: Both 'mode' and 'file' flags are required.")
		flag.Usage()
		os.Exit(1)
	}

	if *encryptFlag && *key == "" {
		fmt.Println("Error: 'key' flag is required when encryption is enabled.")
		flag.Usage()
		os.Exit(1)
	}

	switch *mode {
	case "server":
		listener, err := net.Listen("tcp", *addr)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		file, err := os.Create(*filePath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		if *compress {
			decompressedReader, err := gzip.NewReader(conn)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			_, err = io.Copy(file, decompressedReader)
		} else {
			_, err = io.Copy(file, conn)
		}

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if *validate {
			hash := sha256.New()
			_, err = file.Seek(0, io.SeekStart)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			_, err = io.Copy(hash, file)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("SHA-256: %x\n", hash.Sum(nil))
		}

	case "client":
		conn, err := net.Dial("tcp", *addr)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()

		file, err := os.Open(*filePath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		if *validate {
			hash := sha256.New()
			_, err = io.Copy(hash, file)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("SHA-256: %x\n", hash.Sum(nil))
			_, err = file.Seek(0, io.SeekStart)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}

		var finalReader io.Reader = file

		if *compress {
			compressedFile, err := compressFile(file)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			defer compressedFile.Close()
			finalReader = compressedFile
		}

		_, err = io.Copy(conn, finalReader)

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Error: Invalid mode '%s'\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
}
