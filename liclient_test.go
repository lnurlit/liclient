package liclient

import (
	"flag"
	"testing"
)

var client Client

func TestMain(main *testing.M) {
	var secret string

	flag.StringVar(&secret, "test-secret", "", "")
	flag.Parse()

	if secret == "" {
		panic("--test-secret is required")
	}

	var err error

	client, err = New(secret)

	if err != nil {
		panic(err)
	}

	main.Run()
}

func TestCreateWithdrawal(test *testing.T) {
	withdrawal, err := client.CreateWithdrawal(50, "Hello World", "")

	if err != nil {
		test.Fatal(err)
	}

	test.Logf("withdrawal: %+v\n", withdrawal)
}

func TestGetWithdrawal(test *testing.T) {
	created, _ := client.CreateWithdrawal(50, "Hello World", "")

	withdrawal, err := client.GetWithdrawal(created.ID)

	if err != nil {
		test.Fatal(err)
	}

	test.Logf("withdrawal: %+v\n", withdrawal)
}

func TestDeleteWithdrawal(test *testing.T) {
	created, _ := client.CreateWithdrawal(50, "Hello World", "")

	err := client.DeleteWithdrawal(created.ID)

	if err != nil {
		test.Fatal(err)
	}
}
