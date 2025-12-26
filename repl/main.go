package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pion/webrtc/v3"
)

func main() {
	var code string
	fmt.Print("Enter code: ")
	fmt.Scan(&code)

	repl := NewRepl(code, DefaultRendezvous, 60, onConnect, onDisconnect, onMessageReceived)
	if err := repl.Connect(); err != nil {
		panic(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

func onDisconnect(graceful bool) {
	// TODO: We gotta do nothing for now
	fmt.Println("Companion disconnected.")
}

func onConnect(c *webrtc.DataChannel) {
	fmt.Println("Companion connected!")
	testYail := "(begin (require <com.google.youngandroid.runtime>) (process-repl-input -1 (begin (define-syntax protect-enum (lambda (x) (syntax-case x () ((_ enum-value number-value) (if (< com.google.appinventor.components.common.YaVersion:BLOCKS_LANGUAGE_VERSION 34) #'number-value #'enum-value)))))(clear-current-form)))) (begin (require <com.google.youngandroid.runtime>) (process-repl-input -1 (begin (try-catch (let ((attempt (delay (set-form-name \"Screen1\")))) (force attempt)) (exception java.lang.Throwable 'notfound))))) (begin (require <com.google.youngandroid.runtime>) (process-repl-input -1 (begin (call-Initialize-of-components 'Screen1 'Button1 'Button2))))"
	c.SendText(testYail)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Press Enter to read hi.mist and send...")

		for {
			fmt.Print("> ")
			reader.ReadString('\n')
			fmt.Println("\n---------")

			codeBytes, err := os.ReadFile("/home/ekina/GolandProjects/FalconMain/testing/hi.mist")
			if err != nil {
				fmt.Printf("Error reading hi.mist: %v\n", err)
				continue
			}
			inputCode := string(codeBytes)

			// // Pipeline: Mist -> AST -> Blockly XML -> YAIL

			// // 1. Lexing
			// codeContext := &context.CodeContext{SourceCode: &inputCode, FileName: "hi.mist"}
			// tokens := lex.NewLexer(codeContext).Lex()

			// // 2. Parsers
			// langParser := mistparser.NewLangParser(false, tokens)
			// expressions := langParser.ParseAll()
			// for _, e := range expressions {
			// 	fmt.Printf("%+v\n", e)
			// }
			// println()
			// // 3. To Blockly
			// blocks := make([]ast.Block, len(expressions))
			// for i, expression := range expressions {
			// 	blocks[i] = expression.Blockly(true)
			// }

			// xmlBlock := ast.XmlRoot{
			// 	Blocks: blocks,
			// 	XMLNS:  "https://developers.google.com/blockly/xml",
			// }

			// xmlBytes, err := xml.Marshal(xmlBlock)
			// if err != nil {
			// 	fmt.Printf("Error generating XML: %v\n", err)
			// 	continue
			// }
			// indent, err := xml.MarshalIndent(xmlBlock, "", "  ")
			// if err != nil {
			// 	fmt.Printf("Error generating XML: %v\n", err)
			// 	continue
			// }
			// fmt.Printf("Generated XML: \n%s\n", string(indent))

			// // 4. To YAIL
			// yailCode := blocklytoyail.NewParser(string(xmlBytes)).GenerateYAIL()
			yailCode := "(begin   (require <com.google.youngandroid.runtime>)   (process-repl-input -1     (begin " + inputCode + "     )))"

			// 5. Send
			fmt.Printf("Sending YAIL: %s\n", yailCode)
			c.SendText(yailCode)
		}
	}()
}

func onMessageReceived(msg webrtc.DataChannelMessage) {
	type ReplValue struct {
		Status  string `json:"status"`
		Type    string `json:"type"`
		Value   string `json:"value"`
		BlockID string `json:"blockid"`
	}

	type ReplResponse struct {
		Status string      `json:"status"`
		Values []ReplValue `json:"values"`
	}

	var response ReplResponse
	if err := json.Unmarshal(msg.Data, &response); err != nil {
		fmt.Printf("Message received (raw): %s\n", msg.Data)
		return
	}

	if response.Status == "OK" {
		for _, val := range response.Values {
			if val.Status == "OK" {
				if val.Value != "" && val.Value != "*nothing*" {
					fmt.Printf("Result: %s\n", val.Value)
				}
			} else {
				fmt.Printf("Error: %s\n", val.Value)
			}
		}
	} else {
		fmt.Printf("Repl Error: %s\n", string(msg.Data))
	}
}
