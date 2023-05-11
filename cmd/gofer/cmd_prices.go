//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/chronicleprotocol/oracle-suite/pkg/config"
	"github.com/chronicleprotocol/oracle-suite/pkg/price/provider"
)

func NewPricesCmd(opts *options) *cobra.Command {
	return &cobra.Command{
		Use:     "prices [PAIR...]",
		Aliases: []string{"price"},
		Args:    cobra.MinimumNArgs(0),
		Short:   "Return prices for given PAIRs",
		Long:    `Return prices for given PAIRs.`,
		RunE: func(c *cobra.Command, args []string) (err error) {
			if err := config.LoadFiles(&opts.Config, opts.ConfigFilePath); err != nil {
				return err
			}
			ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt)
			services, err := opts.Config.ClientServices(ctx, opts.Logger(), opts.NoRPC, opts.Format.format)
			if err != nil {
				return err
			}
			if err = services.Start(ctx); err != nil {
				return err
			}
			defer func() {
				if err != nil {
					exitCode = 1
					_ = services.Marshaller.Write(os.Stderr, err)
				}
				_ = services.Marshaller.Flush()
				// Set err to nil because error was already handled by marshaller.
				err = nil
			}()
			defer func() {
				ctxCancel()
				if sErr := <-services.Wait(); err == nil { // Ignore sErr if another error has already occurred.
					err = sErr
				}
			}()
			pairs, err := provider.NewPairs(args...)
			if err != nil {
				return err
			}
			prices, err := services.PriceProvider.Prices(pairs...)
			if err != nil {
				return err
			}
			err = services.PriceHook.Check(prices)
			if err != nil {
				return err
			}
			for _, p := range prices {
				if mErr := services.Marshaller.Write(os.Stdout, p); mErr != nil {
					_ = services.Marshaller.Write(os.Stderr, mErr)
				}
			}
			// If any pair has been returned with an error, then we should return a non-zero status code.
			for _, p := range prices {
				if p.Error != "" {
					exitCode = 1
					break
				}
			}
			return
		},
	}
}
