package main

import (
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"

	steam "github.com/Philipp15b/go-steam"
	"github.com/pkg/errors"
	"github.com/playnet-public/flagenv"
)

var (
	steamAPIKey       = flagenv.String("apiKey", "", "The Steam Web-API key needed to fetch specific information.")
	steamLogin        = flagenv.String("login", "", "The Steam user name of the trader account.")
	steamPass         = flagenv.String("password", "", "The Steam password of the trader account.")
	steamSharedSecret = flagenv.String("sharedSecret", "", "The shared secret of a Steam account.")
	session           *steam.Session
	processTimeout    = flagenv.Int("processTimeout", 60, "The timeout to continue processing all offers.")
	logger            *zap.Logger
)

func main() {
	flagenv.Parse()

	l, err := zap.NewProductionConfig().Build()
	if err != nil {
		log.Fatalln(err.Error())
	}
	logger = l

	logger.Info("starting new Steam session")
	session = steam.NewSessionWithAPIKey(*steamAPIKey)
	if err := login(); err != nil {
		logger.Fatal(err.Error())
	}

	logger.Info("starting trade offer loop")
	for {
		if err := processActiveOffers(); err != nil {
			logger.Fatal(err.Error())
		}
		time.Sleep(time.Duration(*processTimeout) * time.Second)
	}
}

func getTimeDiff() (time.Duration, error) {
	timeTip, err := steam.GetTimeTip()
	if err != nil {
		return 0, err
	}

	return time.Duration(timeTip.Time - time.Now().Unix()), nil
}

func login() error {
	timeDiff, err := getTimeDiff()
	if err != nil {
		return err
	}

	return errors.Wrap(session.Login(*steamLogin, *steamPass, *steamSharedSecret, timeDiff), "steam login")
}

func processActiveOffers() error {
	logger.Info("fetching tradeoffers")
	tOffers, err := session.GetTradeOffers(
		steam.TradeFilterRecvOffers|steam.TradeFilterActiveOnly,
		time.Now(),
	)

	if err != nil {
		return err
	}

	for _, v := range tOffers.Descriptions {
		fmt.Println(v.Name)
	}

	logger.Info(fmt.Sprintf("fetched %d tradeoffers", len(tOffers.ReceivedOffers)))

	for _, v := range tOffers.ReceivedOffers {
		var partnerSID steam.SteamID
		partnerSID.ParseDefaults(v.Partner)

		l := logger.With(
			zap.Uint64("offer_id", v.ID),
			zap.Uint64("partner_steamid64", uint64(partnerSID)),
			zap.Uint8("state", v.State),
		)

		l.Info("checking offer")
		if v.State == steam.TradeStateActive {
			if len(v.SendItems) > 0 {
				l.Info("cancel offer: no items received!")
				if err := v.Cancel(session); err != nil {
					return err
				}
			} else {
				l.Info("accepting offer")
				if err := v.Accept(session); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
