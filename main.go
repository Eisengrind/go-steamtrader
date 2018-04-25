package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/doctype/steam"
	"github.com/playnet-public/flagenv"
)

var (
	STEAM_API_KEY       = ""
	STEAM_LOGIN         = ""
	STEAM_PASS          = ""
	STEAM_SHARED_SECRET = ""
	SESSION             *steam.Session
	PROCESSTIMEOUT      = 60
)

func main() {
	flagenv.EnvPrefix = "steamtrader"
	flagenv.StringVar(&STEAM_API_KEY, "apiKey", "", "The Steam Web-API key needed to fetch specific information.")
	flagenv.StringVar(&STEAM_LOGIN, "login", "", "The Steam user name of the trader account.")
	flagenv.StringVar(&STEAM_PASS, "password", "", "The Steam password of the trader account.")
	flagenv.StringVar(&STEAM_SHARED_SECRET, "sharedSecret", "", "The shared secret of a Steam account.")
	flagenv.IntVar(&PROCESSTIMEOUT, "processTimeout", 60, "The timeout to continue processing all offers.")
	flagenv.Parse()

	if err := checkVars(); err != nil {
		log.Fatalln(err.Error())
	}

	SESSION = newSession()
	if err := login(); err != nil {
		log.Fatalln(err.Error())
	}

	for {
		if err := processActiveOffers(); err != nil {
			log.Println(err.Error())
		}
		time.Sleep(time.Duration(PROCESSTIMEOUT) * time.Second)
	}
}

var (
	ErrSteamApiKeyEmpty       = errors.New("given Steam api key is empty")
	ErrSteamLoginEmpty        = errors.New("given Steam login name is empty")
	ErrSteamPassEmpty         = errors.New("given Steam password is empty")
	ErrSteamSharedSecretEmpty = errors.New("given Steam shared secret is empty")
	ErrTimeoutNotCorrect      = errors.New("given timeout is equal-less than 0")
)

func checkVars() error {
	if STEAM_API_KEY == "" {
		return ErrSteamApiKeyEmpty
	}

	if STEAM_LOGIN == "" {
		return ErrSteamLoginEmpty
	}

	if STEAM_PASS == "" {
		return ErrSteamPassEmpty
	}

	if STEAM_SHARED_SECRET == "" {
		return ErrSteamSharedSecretEmpty
	}

	if PROCESSTIMEOUT <= 0 {
		return ErrTimeoutNotCorrect
	}

	return nil
}

func getTimeDiff() (time.Duration, error) {
	timeTip, err := steam.GetTimeTip()
	if err != nil {
		return 0, err
	}

	return time.Duration(timeTip.Time - time.Now().Unix()), nil
}

func newSession() *steam.Session {
	return steam.NewSessionWithAPIKey(STEAM_API_KEY)
}

func login() error {
	timeDiff, err := getTimeDiff()
	if err != nil {
		return err
	}

	return SESSION.Login(STEAM_LOGIN, STEAM_PASS, STEAM_SHARED_SECRET, timeDiff)
}

func processActiveOffers() error {
	tOffers, err := SESSION.GetTradeOffers(
		steam.TradeFilterRecvOffers|steam.TradeFilterActiveOnly,
		time.Now(),
	)

	if err != nil {
		return err
	}

	for _, v := range tOffers.Descriptions {
		fmt.Println(v.Name)
	}

	for _, v := range tOffers.ReceivedOffers {
		if v.State == steam.TradeStateActive {
			if len(v.SendItems) > 0 {
				if err := v.Cancel(SESSION); err != nil {
					return err
				}
			} else {
				if err := v.Accept(SESSION); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
