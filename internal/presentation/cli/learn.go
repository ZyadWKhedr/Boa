package cli

import (
	"fmt"
	"strings"

	"compressor/internal/domain"
	"compressor/internal/ui"
	"github.com/spf13/cobra"
)

func (a *App) newLearnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "learn [technique-id]",
		Aliases: []string{"explain", "guide", "edu"},
		Short:   "Educational guide explaining how compression techniques work",
		Long: `Provides clear, plain-language explanations of compression techniques,
analogies, tradeoffs (speed vs size vs compatibility), and deep technical principles.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return a.renderLearnCatalog()
			}
			return a.renderLearnDetail(args[0])
		},
	}

	return cmd
}

func (a *App) renderLearnCatalog() error {
	if !flagQuiet {
		ui.PrintBanner()
		fmt.Println()
	}

	ui.PrintSection("Boa Compression Knowledge Base 📚")
	fmt.Println(" Learn how compression algorithms work, what they trade off, and when to use them.")
	fmt.Println()

	techniques := a.LearnUC.ListTechniques()

	var lossless []domain.TechniqueInfo
	var lossy []domain.TechniqueInfo

	for _, t := range techniques {
		if t.Category == domain.CategoryLossy {
			lossy = append(lossy, t)
		} else {
			lossless = append(lossless, t)
		}
	}

	fmt.Println(ui.Bold(ui.Green("▶ Lossless Techniques (Bit-for-bit exact recovery)")))
	for _, t := range lossless {
		fmt.Printf("   %-16s %s\n", ui.Bold(ui.Cyan(t.ID)), t.Name)
		fmt.Printf("   %-16s %s\n\n", "", ui.Dim(t.Summary))
	}

	fmt.Println(ui.Bold(ui.Yellow("▶ Lossy Techniques (Perceptual optimization for media)")))
	for _, t := range lossy {
		fmt.Printf("   %-16s %s\n", ui.Bold(ui.Yellow(t.ID)), t.Name)
		fmt.Printf("   %-16s %s\n\n", "", ui.Dim(t.Summary))
	}

	fmt.Printf(" Run %s to dive into full details, analogies, and technical secrets!\n\n",
		ui.Bold(ui.Cyan("boa learn <technique-id>")))

	return nil
}

func (a *App) renderLearnDetail(id string) error {
	info, err := a.LearnUC.GetTechniqueInfo(strings.ToLower(strings.TrimSpace(id)))
	if err != nil {
		return err
	}

	if !flagQuiet {
		ui.PrintBanner()
		fmt.Println()
	}

	categoryBadge := ui.BadgeOK("LOSSLESS")
	if info.Category == domain.CategoryLossy {
		categoryBadge = ui.BadgeWarn("LOSSY MEDIA")
	}

	fmt.Printf(" %s  %s\n", categoryBadge, ui.Bold(ui.Cyan(info.Name)))
	fmt.Printf(" %s\n\n", ui.Dim(info.Summary))

	fmt.Printf(" %s %s\n\n", ui.Bold("💡 Analogy:"), info.Analogy)

	fmt.Printf(" %s %s\n", ui.Bold(ui.Green("✔ What You Gain:")), info.WhatYouGain)
	fmt.Printf(" %s %s\n\n", ui.Bold(ui.Red("✖ What You Lose:")), info.WhatYouLose)

	fmt.Printf(" %s %s\n", ui.Bold("🎯 Best For:"), info.BestFor)
	fmt.Printf(" %s %s\n\n", ui.Bold("⚠️ Avoid For:"), info.AvoidFor)

	if info.GoDeeper != "" {
		fmt.Printf(" %s\n", ui.Bold(ui.Magenta("🔬 Go Deeper (Technical Principles):")))
		fmt.Printf("   %s\n\n", ui.Dim(info.GoDeeper))
	}

	return nil
}
