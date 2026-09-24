package cmd

import (
	"fmt"

	"github.com/hungovercoders/terminal-of-terror/internal/crypt"
	"github.com/hungovercoders/terminal-of-terror/internal/monsters"
	"github.com/hungovercoders/terminal-of-terror/internal/quiz"
	"github.com/hungovercoders/terminal-of-terror/internal/ui"
	"github.com/spf13/cobra"
)

var (
	quizQuestions int
	quizJSON      bool
)

var quizCmd = &cobra.Command{
	Use:   "quiz [monster]",
	Short: "Test your monster knowledge in the Midnight Quiz",
	Long: `Answer multiple-choice questions about facts, films, quotes and myths.

Every correct answer about a monster counts towards capturing it for your
crypt. Name a monster to be quizzed only about that one.`,
	Example: `  terminal-of-terror quiz
  terminal-of-terror quiz --questions 5
  terminal-of-terror quiz dracula`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if quizQuestions < 1 {
			return fmt.Errorf("--questions must be at least 1")
		}
		focus := ""
		if len(args) == 1 {
			m, err := resolveMonster(args[0])
			if err != nil {
				return err
			}
			focus = m.ID
		}
		r := newRand()
		qs := quiz.Generate(r, monsters.GetAllMonsters(), quizQuestions, focus)
		if len(qs) == 0 {
			return fmt.Errorf("couldn't find any questions to ask")
		}
		if quizJSON {
			return printJSON(qs)
		}
		res, err := ui.RunQuiz(qs, r)
		if err != nil {
			return err
		}
		_, answered := res.Score()
		if answered == 0 {
			return nil
		}
		record(crypt.Outcome{
			Kind:      "quiz",
			Correct:   res.CorrectBy(),
			Score:     res.Percent(),
			Total:     len(qs),
			Completed: res.Completed,
		})
		return nil
	},
}

func init() {
	rootCmd.AddCommand(quizCmd)
	quizCmd.Flags().IntVarP(&quizQuestions, "questions", "n", 10, "Number of questions")
	quizCmd.Flags().BoolVar(&quizJSON, "json", false, "Print the questions and answers as JSON instead of playing")
}
