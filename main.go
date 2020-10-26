package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strconv"

	"github.com/BurntSushi/toml"
	goton "github.com/move-ton/ton-client-go"
	crypto "github.com/move-ton/ton-client-go/crypto"
	tvm "github.com/move-ton/ton-client-go/tvm"
	"github.com/tealeg/xlsx/v3"
)

var (
	networkChainID, persentRewardJurys        int
	addressContest, filePath, proposalAddress string
)

func init() {
	flag.IntVar(&networkChainID, "chain", 0, "0-mainnet, 1-devnet")
	flag.StringVar(&addressContest, "addr", "", "Address contract contest")
	flag.StringVar(&filePath, "fpath", "", "Way to file with rewards")
	flag.StringVar(&proposalAddress, "propaddr", "", "Proposal addres for make hyperlink submission id first table")
	flag.IntVar(&persentRewardJurys, "prz", 5, "Persent rewards from sum for jurys")
}

func main() {

	flag.Parse()

	nameChain := "main.ton.dev"
	if addressContest == "" {
		log.Fatal("Parameter addr is empty!")
	}

	var configTml TomlConfig

	if filePath == "" {
		fmt.Println("Way to file with rewards is empty, file filled without rewards!")
	} else {
		if _, err := toml.DecodeFile(filePath, &configTml); err != nil {
			fmt.Println("Error read file with rewards: ", err, ", file filled without rewards")
		}
	}

	if networkChainID != 0 {
		nameChain = "net.ton.dev"
	}

	config, err := goton.ParseConfigFile("config.toml")
	if err != nil {
		log.Println("Error read config file, err: . Settings setup on default.", err)
		config = goton.NewConfig(0)
	}

	config.Network.ServerAddress = nameChain
	client, err := goton.InitClient(config)
	if err != nil {
		log.Fatal("Init client error", err)
	}
	defer client.Destroy()

	file, err := os.Open("FreeTonContest.abi.json")
	if err != nil {
		fmt.Println("Error 1 open: ", err)
		return
	}

	abiByte, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Error 2 read file abi: ", err)
		return
	}

	keys, _ := crypto.GenerateRandomSignKeysResult(client.Request(crypto.GenerateRandomSignKeys()))
	dd := &tvm.ParamsOfExecuteMessage{}
	dd.Message = &tvm.MessageSourceEncParam{Type: "EncodingParams", Abi: tvm.Abi{Type: "Serialized", Value: abiByte}, Address: "0:8fcfffc70abd7bc6fbfd4d2d56b7712bef5fbb9d6fba057de0d5539dcbbc71e5", Signer: tvm.GetSignerKeys(keys), CallSet: &tvm.CallSet{FunctionName: "listContenders"}}
	dd.Account = "te6ccgICAS8AAQAANhYAAAJ1wAj8//xwq9e8b7/U0tVrdxK+9fu51vugV94NVTncu8ceVAJewGFZQvx4UQAAABblyMnzGUDK1+JcE0AAagABBO0AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMABAAAAXVGv4nNgAAAAYACQAlQ3US3De7+atVFkZkrnCHEPG1XObs/eO0bsCy/6f/AqABxiBRVXi3A6g7/xPEmh2osPHfAy+2GUbVv6oMhpsmwSIABAEAAIABBAAYAAwACACBfhvfRX4QiGF+E7f9fjih/AkgAAwAEzcPaZt1q6wObzGpaVYpp2rPFzudH+KX+MqPMKBgTdbAABQAEAJJodHRwczovL2ZvcnVtLmZyZWV0b24ub3JnL3QvY29udGVzdC1mcmVlLXRvbi1kZWZpLWp1cnktc2VsZWN0aW9uLTEtMS8yODY5AEBGcmVlIFRPTiBEZUZpIEp1cnkgU2VsZWN0aW9uIDEuMQMB6ABAADIABwICzgAkAAgCASAAEgAJAQEgAAoCAZAACwAUAiHAAMAAwAAAAAAAAAAAAAAHoAAMABkCAs0ADQAdAgEgABAADgEBWAAPAC52b3RlIHRvIGFjY2VwdCBmb3IganVyeQEBWAARACZWZXJ5IGdvb2QgY2FuZGlkYXRlAQEgABMCAZAAGAAUAgLNABYAFQAJaAAAACoCASAAFwAXAAlQAAAAqAIhwADAAMAAAAAAAAAAAAAAB6AAHAAZAgLNABoAKwIBIAAbABsAA1AYAgLNAB8AHQEBagAeAGBWZXJ5IGdvb2QgYW5kIGNvbXBldGl0aXZlIGp1cnkgY2FuZGlkYXRlICh2YXRpYykCASAAIgAgAQFYACEAGlZvdGUgZm9yIGp1cnkBAVgAIwBQVmVyeSBnb29kIGFuZCBjb21wZXRpdGl2ZSBqdXJ5IGNhbmRpZGF0ZQEBWAAlAgEwACkAJgICzQAoACcAAWsAAWcCIcAAgAAAAAAAgAAAAAAAAAAgAC0AKgICzQAsACsAA2gGAANkBgICzQAwAC4BAWoALwAcZG91YmxlICh2YXRpYykBAWYAMQAkSW52YWxpZCBzdWJtaXNzaW9uAgLOADwAMwIBIAA5ADQC1iABRDDU6lTkuzcIlPuZHwDdBkxhnqoKYAG1wMoUF/vV4uGk4dew87twKARnTg9GI9HoneHAtpeCeAWoGijT67oS4AAAAAL8JJ7MAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADcANQH+aHR0cHM6Ly9maXJlYmFzZXN0b3JhZ2UuZ29vZ2xlYXBpcy5jb20vdjAvYi90b24tbGFicy5hcHBzcG90LmNvbS9vL2RvY3VtZW50cyUyRmFwcGxpY2F0aW9uJTJGcGRmJTJGZzhraTRuZThyemNrZzZ0aGRsbC1GcmVlVE9OJQA2AJgyMGp1cnklMjBzdWJtaXNzaW9uLnBkZj9hbHQ9bWVkaWEmdG9rZW49MGZlYzk4YWQtOTg5OC00NGIzLTg0Y2EtMzJmM2RjZTllODJlAf5odHRwczovL2ZvcnVtLmZyZWV0b24ub3JnL3N1Ym1pc3Npb24/cHJvcG9zYWxBZGRyZXNzPTA6MmE1YWMxMjRiMThmMmVkZWU5NGUzMTU2OWViNmEwMDk0OGEzNzk3ZmRmYTNiMWEzMTY4ZDU3OWRhMmZiMDIxZiZzdWJtaXNzADgAEGlvbklkPTEyAtYgBPcq9E4dpDAemmsvs/JjHEMdDRyQW87rMWvmxiQnu6Iq7fvszx2FFsI3gUO8NQmNULIiczbZweHZwzwOqVChR1AAAAAC/CFh/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA/ADoB/mh0dHBzOi8vZmlyZWJhc2VzdG9yYWdlLmdvb2dsZWFwaXMuY29tL3YwL2IvdG9uLWxhYnMuYXBwc3BvdC5jb20vby9kb2N1bWVudHMlMkZhcHBsaWNhdGlvbiUyRnBkZiUyRmR0d2ppcDd6OGlma2c2ZG95ZnMtSWduYXQlMjAAOwCWU2hhcGtpbl9zdWJtaXNzaW9uLnBkZj9hbHQ9bWVkaWEmdG9rZW49ZjU1YmE2YzgtOGE3ZS00MGIzLWE4NTAtZDA2NTcyYzVmMjJlAtdYAT3KvROHaQwHpprL7PyYxxDHQ0ckFvO6zFr5sYkJ7uiKDA+yuGr78HvknaDAUz3raPmu3sNrEuW6NexenasORYwAAAAAvwhVtwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACAAPwA9Af5odHRwczovL2ZpcmViYXNlc3RvcmFnZS5nb29nbGVhcGlzLmNvbS92MC9iL3Rv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"

	dd.Mode = "TvmOnly"
	dd.Message.CallSet.Header.PubKey = keys.Public
	val1 := client.RequestAsync(tvm.ExecuteMessage(dd))
	if err != nil {
		log.Fatal("Error get version, err: ", err)
		return
	}

	res, err := tvm.ExecuteMessageResult(client.GetResp(val1))
	if err != nil {
		fmt.Println("Error result: ", err)
		return
	}

	md := &mainDats{}
	res1 := &resultContenders{}
	_ = json.Unmarshal(res.Decoded.Output, res1)

	lenC := len(res1.Addresses)
	for i := 0; i < lenC; i++ {
		contDrs := Contenders{}
		contDrs.IDS, _ = strconv.ParseInt(res1.Ids[i], 0, 64)
		contDrs.Address = res1.Addresses[i]

		md.Contenders = append(md.Contenders, contDrs)
	}

	dd.Message.CallSet.FunctionName = "getContestInfo"
	val2 := client.RequestAsync(tvm.ExecuteMessage(dd))
	if err != nil {
		log.Fatal("Error get version, err: ", err)
		return
	}

	res, err = tvm.ExecuteMessageResult(client.GetResp(val2))
	if err != nil {
		fmt.Println("Error result: ", err)
		return
	}

	res2 := &resContestInfo{}
	_ = json.Unmarshal(res.Decoded.Output, res2)

	md.TitleContext = hexToString([]byte(res2.Title))
	md.LinkToContext = hexToString([]byte(res2.Link))
	lenJ := len(res2.JuryKeys)
	for i := 0; i < lenJ; i++ {
		juryS := Jury{}
		juryS.Address = res2.JuryAddresses[i]
		juryS.PublicKey = res2.JuryKeys[i]

		md.Jurys = append(md.Jurys, juryS)
	}

	mm := make(map[string]votes)
	dd.Message.CallSet.FunctionName = "getVotesPerJuror"

	//Асинхронная обработка?!
	for n, val := range md.Contenders {
		idReq := req{ID: val.IDS}
		dd.Message.CallSet.Input = idReq

		val3 := client.RequestAsync(tvm.ExecuteMessage(dd))
		if err != nil {
			log.Fatal("Error get version, err: ", err)
			return
		}

		fmt.Println(n, " ", val3)

		res, err = tvm.ExecuteMessageResult(client.GetResp(val3))
		if err != nil {
			fmt.Println("Error result: ", err)
			return
		}

		res3 := &goverment{}
		_ = json.Unmarshal(res.Decoded.Output, res3)

		md.Contenders[n].GovermentD = res3

		for _, valFor := range res3.JurorsFor {
			if vvF, found := mm[valFor]; found {
				vvF.JuryFor++
				mm[valFor] = vvF
			} else {
				mm[valFor] = votes{1, 0, 0}
			}
		}

		for _, valAbs := range res3.JurorsAbstained {
			if vvAb, found := mm[valAbs]; found {
				vvAb.JuryAbstained++
				mm[valAbs] = vvAb
			} else {
				mm[valAbs] = votes{0, 1, 0}
			}
		}

		for _, valAg := range res3.JurorsAgainst {
			if vvAg, found := mm[valAg]; found {
				vvAg.JuryAgainst++
				mm[valAg] = vvAg
			} else {
				mm[valAg] = votes{0, 0, 1}
			}
		}

		var totalFF, sumCount int64
		for _, valMarks := range res3.Marks {
			ff, err := strconv.ParseInt(valMarks, 0, 16)
			if err != nil {
				continue
			}

			totalFF += ff
			sumCount++
		}

		md.Contenders[n].AverageScore = float64(totalFF) / float64(sumCount)
		md.Contenders[n].Reject = int64(len(res3.JurorsAgainst))
	}

	// sort.SliceStable(md.Contenders, func(i, j int) bool {
	// 	return md.Contenders[i].AverageScore > md.Contenders[j].AverageScore
	// })
	sort.Slice(md.Contenders, func(i, j int) bool {
		return md.Contenders[i].AverageScore > md.Contenders[j].AverageScore
	})

	var (
		intprev float64
		place   int64
	)

	place = 1
	type strT struct {
		Count  int64
		Reward int64
	}
	m1 := make(map[int64]*strT)
	for valN, valueC := range md.Contenders {

		countVotest := float64(int64(len(valueC.GovermentD.JurorsFor) + len(valueC.GovermentD.JurorsAbstained) + len(valueC.GovermentD.JurorsAgainst)))
		if math.IsNaN(valueC.AverageScore) || valueC.AverageScore == 0 || (float64(valueC.Reject)/countVotest)*100 > 50 {
			continue
		}

		if place == 1 {
			intprev = valueC.AverageScore
			*(&md.Contenders[valN].Ranking) = place
			// .valueC.Ranking = place
			if v, found := m1[place]; !found {
				m1[place] = &strT{Count: 1, Reward: 0}
			} else {
				v.Count++
			}
		} else {
			if valueC.AverageScore < intprev {
				intprev = valueC.AverageScore
				*(&md.Contenders[valN].Ranking) = place
				// valueC.Ranking = place
				if v, found := m1[place]; !found {
					m1[place] = &strT{Count: 1, Reward: 0}
				} else {
					v.Count++
				}
			} else {
				*(&md.Contenders[valN].Ranking) = place - 1
				// valueC.Ranking = place - 1
				if v, found := m1[place-1]; !found {
					m1[place-1] = &strT{Count: 1, Reward: 0}
				} else {
					v.Count++
				}
			}
		}
		place++
	}

	sort.Slice(md.Contenders, func(i, j int) bool {
		return md.Contenders[i].IDS < md.Contenders[j].IDS
	})

	if filePath != "" {
		var allMoney int64
		for v, valM := range m1 {
			if valM.Count != 1 {
				var summNow int64
				for _, valCc := range configTml.Rewards[v-1 : valM.Count] {
					summNow += valCc
				}
				valM.Reward += summNow / int64(len(configTml.Rewards[v-1:valM.Count]))
				allMoney += summNow
			} else {
				valM.Reward = configTml.Rewards[v]
				allMoney += valM.Reward
			}
		}

		md.RewardsSumCont = allMoney

		for _, valCont := range md.Contenders {
			valCont.Reward = m1[valCont.Ranking].Reward
		}
	}

	if err := generateFile(md, mm); err != nil {
		log.Fatal("Error generate file: ", err)
	}

	fmt.Println("Create file: ", md.TitleContext+".xlsx")
}

func generateFile(data *mainDats, mm map[string]votes) error {
	wb := xlsx.NewFile()
	sheet1, _ := wb.AddSheet("Main")
	sheet1.SetColWidth(0, 0, 15)
	sheet1.SetColWidth(1, 1, 70)
	sheet1.SetColWidth(2, 6, 15)

	addEmptyString(sheet1, 0, 0)

	style1 := xlsx.NewStyle()
	style1.Font.Name = "Arial"
	style1.Font.Size = 24

	style2 := xlsx.NewStyle()
	style2.Font.Name = "Arial"
	style2.Font.Size = 10

	row2 := sheet1.AddRow()
	cell1R2 := row2.AddCell()
	cell1R2.SetHyperlink(data.LinkToContext, data.TitleContext, "")
	cell1R2.SetStyle(style1)
	nSt := cell1R2.GetStyle()
	nSt.Font.Color = "1155CC"
	nSt.Font.Bold = true
	nSt.Font.Underline = true

	addEmptyString(sheet1, 2, 0)

	st := xlsx.NewStyle()
	st.Font.Name = "Arial"
	st.Font.Size = 10
	st.Font.Color = "FFFFFF"
	st.Font.Bold = true
	st.Fill.FgColor = "5B95F9"
	st.Fill.PatternType = "solid"
	st.Alignment.Horizontal = "center"

	row4 := sheet1.AddRow()
	cell1R4 := row4.AddCell()
	cell1R4.SetString("Submission №")
	cell1R4.SetStyle(st)

	cell2R4 := row4.AddCell()
	cell2R4.SetString("Wallet Address")
	cell2R4.SetStyle(st)

	cell3R4 := row4.AddCell()
	cell3R4.SetString("Average score")
	cell3R4.SetStyle(st)

	cell4R4 := row4.AddCell()
	cell4R4.SetString("Ranking")
	cell4R4.SetStyle(st)

	cell5R4 := row4.AddCell()
	cell5R4.SetString("Reward")
	cell5R4.SetStyle(st)

	cell6R5 := row4.AddCell()
	cell6R5.SetString("Accepted")
	cell6R5.SetStyle(st)

	cell7R5 := row4.AddCell()
	cell7R5.SetString("Abstained")
	cell7R5.SetStyle(st)

	cell5R5 := row4.AddCell()
	cell5R5.SetString("Rejected")
	cell5R5.SetStyle(st)

	blueColor, propAddressUse := false, false
	nowLinkSub := ""
	if proposalAddress != "" {
		propAddressUse = true
		nowLinkSub = fmt.Sprintf(linkToSubmissionGov, proposalAddress)
	}

	for _, val := range data.Contenders {
		row5 := sheet1.AddRow()
		cell1R5 := row5.AddCell()
		cell1R5.GetStyle().Font.Name = "Arial"
		cell1R5.GetStyle().Font.Size = 10
		cell1R5.GetStyle().Font.Bold = true
		cell1R5.GetStyle().Alignment.Horizontal = "center"
		if !propAddressUse {
			cell1R5.SetValue(val.IDS)
		} else {
			isdS := strconv.FormatInt(val.IDS, 10)
			cell1R5.SetHyperlink(nowLinkSub+isdS, isdS, "")
			cell1R5.GetStyle().Font.Color = "1155CC"
			cell1R5.GetStyle().Font.Underline = true
		}

		if val.Ranking == 0 {
			cell1R5.GetStyle().Fill.FgColor = "F4CCCC"
			cell1R5.GetStyle().Fill.PatternType = "solid"
		} else if blueColor {
			cell1R5.GetStyle().Fill.FgColor = "E8F0FE"
			cell1R5.GetStyle().Fill.PatternType = "solid"
		}

		cell2R5 := row5.AddCell()
		cell2R5.GetStyle().Font.Name = "Arial"
		cell2R5.SetHyperlink(linkToExplorer+val.Address, val.Address, "")
		cell2R5.GetStyle().Font.Size = 10
		cell2R5.GetStyle().Font.Color = "1155CC"
		cell2R5.GetStyle().Font.Underline = true
		if val.Ranking == 0 {
			cell2R5.GetStyle().Fill.FgColor = "F4CCCC"
			cell2R5.GetStyle().Fill.PatternType = "solid"
		} else if blueColor {
			cell2R5.GetStyle().Fill.FgColor = "E8F0FE"
			cell2R5.GetStyle().Fill.PatternType = "solid"
		}

		cell3R5 := row5.AddCell()
		cell3R5.SetFloatWithFormat(val.AverageScore, "#0.00")
		cell3R5.GetStyle().Font.Name = "Arial"
		cell3R5.GetStyle().Font.Size = 10
		cell3R5.GetStyle().Alignment.Horizontal = "center"
		cell3R5.GetStyle().Font.Bold = true
		if val.Ranking == 0 {
			cell3R5.GetStyle().Fill.FgColor = "F4CCCC"
			cell3R5.GetStyle().Fill.PatternType = "solid"
		} else if blueColor {
			cell3R5.GetStyle().Fill.FgColor = "E8F0FE"
			cell3R5.GetStyle().Fill.PatternType = "solid"
		}

		cell4R5 := row5.AddCell()
		cell4R5.GetStyle().Font.Name = "Arial"
		cell4R5.GetStyle().Font.Size = 10
		cell4R5.GetStyle().Font.Bold = true
		cell4R5.SetInt64(val.Ranking)
		cell4R5.GetStyle().Alignment.Horizontal = "center"
		if val.Ranking == 0 {
			cell4R5.GetStyle().Fill.FgColor = "F4CCCC"
			cell4R5.GetStyle().Fill.PatternType = "solid"
		} else if blueColor {
			cell4R5.GetStyle().Fill.FgColor = "E8F0FE"
			cell4R5.GetStyle().Fill.PatternType = "solid"
		}

		cell5R9 := row5.AddCell()
		cell5R9.GetStyle().Font.Name = "Arial"
		cell5R9.GetStyle().Font.Size = 10
		cell5R9.GetStyle().Font.Bold = true
		cell5R9.SetInt64(val.Reward)
		cell5R9.GetStyle().Alignment.Horizontal = "center"
		if val.Ranking == 0 {
			cell5R9.GetStyle().Fill.FgColor = "F4CCCC"
			cell5R9.GetStyle().Fill.PatternType = "solid"
		} else if blueColor {
			cell5R9.GetStyle().Fill.FgColor = "E8F0FE"
			cell5R9.GetStyle().Fill.PatternType = "solid"
		}

		cell7R9 := row5.AddCell()
		cell7R9.SetInt(len(val.GovermentD.JurorsFor))
		cell7R9.GetStyle().Font.Name = "Arial"
		cell7R9.GetStyle().Font.Size = 10
		cell7R9.GetStyle().Alignment.Horizontal = "center"
		cell7R9.GetStyle().Font.Bold = true
		if val.Ranking == 0 {
			cell7R9.GetStyle().Fill.FgColor = "F4CCCC"
			cell7R9.GetStyle().Fill.PatternType = "solid"
		} else if blueColor {
			cell7R9.GetStyle().Fill.FgColor = "E8F0FE"
			cell7R9.GetStyle().Fill.PatternType = "solid"
		}

		cell8R9 := row5.AddCell()
		cell8R9.SetInt(len(val.GovermentD.JurorsAbstained))
		cell8R9.GetStyle().Font.Name = "Arial"
		cell8R9.GetStyle().Font.Size = 10
		cell8R9.GetStyle().Alignment.Horizontal = "center"
		cell8R9.GetStyle().Font.Bold = true
		if val.Ranking == 0 {
			cell8R9.GetStyle().Fill.FgColor = "F4CCCC"
			cell8R9.GetStyle().Fill.PatternType = "solid"
		} else if blueColor {
			cell8R9.GetStyle().Fill.FgColor = "E8F0FE"
			cell8R9.GetStyle().Fill.PatternType = "solid"
		}

		cell6R9 := row5.AddCell()
		cell6R9.SetInt64(val.Reject)
		cell6R9.GetStyle().Font.Name = "Arial"
		cell6R9.GetStyle().Font.Size = 10
		cell6R9.GetStyle().Alignment.Horizontal = "center"
		cell6R9.GetStyle().Font.Bold = true
		if val.Ranking == 0 {
			cell6R9.GetStyle().Fill.FgColor = "F4CCCC"
			cell6R9.GetStyle().Fill.PatternType = "solid"
		} else if blueColor {
			cell6R9.GetStyle().Fill.FgColor = "E8F0FE"
			cell6R9.GetStyle().Fill.PatternType = "solid"
		}

		blueColor = !blueColor
	}

	stylefoo1 := xlsx.NewStyle()
	stylefoo1.Font.Name = "Arial"
	stylefoo1.Font.Size = 10
	stylefoo1.Font.Bold = true
	stylefoo1.Alignment.Horizontal = "center"
	stylefoo1.Fill.FgColor = "ACC9FE"
	stylefoo1.Fill.PatternType = "solid"

	row45 := sheet1.AddRow()
	cell1R45 := row45.AddCell()
	cell1R45.SetString(" ")
	cell1R45.SetStyle(stylefoo1)

	cell2R45 := row45.AddCell()
	cell2R45.SetString(" ")
	cell2R45.SetStyle(stylefoo1)

	cell3R45 := row45.AddCell()
	cell3R45.SetString(" ")
	cell3R45.SetStyle(stylefoo1)

	cell4R45 := row45.AddCell()
	cell4R45.SetString("Total:")
	cell4R45.SetStyle(stylefoo1)

	cell5R45 := row45.AddCell()
	cell5R45.SetInt64(data.RewardsSumCont)
	cell5R45.SetStyle(stylefoo1)

	cell6R45 := row45.AddCell()
	cell6R45.SetString(" ")
	cell6R45.SetStyle(stylefoo1)

	cell7R45 := row45.AddCell()
	cell7R45.SetString(" ")
	cell7R45.SetStyle(stylefoo1)

	cell8R45 := row45.AddCell()
	cell8R45.SetString(" ")
	cell8R45.SetStyle(stylefoo1)

	addEmptyString(sheet1, 2, 0)

	row6 := sheet1.AddRow()
	cell1R6 := row6.AddCell()
	cell1R6.SetString("Jury Rewards")
	cell1R6.SetStyle(style1)
	cell1R6.GetStyle().Font.Bold = true

	addEmptyString(sheet1, 2, 0)

	row8 := sheet1.AddRow()
	cell1R8 := row8.AddCell()
	cell1R8.SetString("Jury №")
	cell1R8.SetStyle(st)

	cell2R8 := row8.AddCell()
	cell2R8.SetString("Wallet Address")
	cell2R8.SetStyle(st)

	cell3R8 := row8.AddCell()
	cell3R8.SetString("Votes count")
	cell3R8.SetStyle(st)

	cell4R8 := row8.AddCell()
	cell4R8.SetString("Reward")
	cell4R8.SetStyle(st)

	cell5R8 := row8.AddCell()
	cell5R8.SetString("Accepted")
	cell5R8.SetStyle(st)

	cell6R8 := row8.AddCell()
	cell6R8.SetString("Abstained")
	cell6R8.SetStyle(st)

	cell7R8 := row8.AddCell()
	cell7R8.SetString("Rejected")
	cell7R8.SetStyle(st)

	indJury := 1
	var countVote, countFor, countAbstained, countAgainst int64
	sumReward := 0.0
	blueColor = false
	for _, valJ := range data.Jurys {
		sumVotes := mm[valJ.Address].JuryFor + mm[valJ.Address].JuryAbstained + mm[valJ.Address].JuryAgainst
		if sumVotes > 0 {
			row9 := sheet1.AddRow()
			cell1R9 := row9.AddCell()
			cell1R9.SetValue(indJury)
			cell1R9.GetStyle().Font.Name = "Arial"
			cell1R9.GetStyle().Font.Size = 10
			cell1R9.GetStyle().Font.Bold = true
			cell1R9.GetStyle().Alignment.Horizontal = "center"
			if blueColor {
				cell1R9.GetStyle().Fill.FgColor = "E8F0FE"
				cell1R9.GetStyle().Fill.PatternType = "solid"
			}

			cell2R9 := row9.AddCell()
			cell2R9.SetValue(valJ.Address)
			cell2R9.GetStyle().Font.Name = "Arial"
			cell2R9.GetStyle().Font.Size = 10
			if blueColor {
				cell2R9.GetStyle().Fill.FgColor = "E8F0FE"
				cell2R9.GetStyle().Fill.PatternType = "solid"
			}

			cell3R9 := row9.AddCell()
			cell3R9.SetValue(sumVotes)
			cell3R9.GetStyle().Font.Name = "Arial"
			cell3R9.GetStyle().Font.Size = 10
			cell3R9.GetStyle().Font.Bold = true
			cell3R9.GetStyle().Alignment.Horizontal = "center"
			if blueColor {
				cell3R9.GetStyle().Fill.FgColor = "E8F0FE"
				cell3R9.GetStyle().Fill.PatternType = "solid"
			}

			cell4R9 := row9.AddCell()
			cell4R9.GetStyle().Font.Name = "Arial"
			cell4R9.GetStyle().Font.Size = 10
			cell4R9.GetStyle().Font.Bold = true
			cell4R9.SetFloatWithFormat(0, "#0.00")
			cell4R9.GetStyle().Alignment.Horizontal = "center"
			if blueColor {
				cell4R9.GetStyle().Fill.FgColor = "E8F0FE"
				cell4R9.GetStyle().Fill.PatternType = "solid"
			}

			cell5R9 := row9.AddCell()
			cell5R9.SetValue(mm[valJ.Address].JuryFor)
			cell5R9.GetStyle().Font.Name = "Arial"
			cell5R9.GetStyle().Font.Size = 10
			cell5R9.GetStyle().Font.Bold = true
			cell5R9.GetStyle().Alignment.Horizontal = "center"
			if blueColor {
				cell5R9.GetStyle().Fill.FgColor = "E8F0FE"
				cell5R9.GetStyle().Fill.PatternType = "solid"
			}

			cell6R9 := row9.AddCell()
			cell6R9.SetValue(mm[valJ.Address].JuryAbstained)
			cell6R9.GetStyle().Font.Name = "Arial"
			cell6R9.GetStyle().Font.Size = 10
			cell6R9.GetStyle().Font.Bold = true
			cell6R9.GetStyle().Alignment.Horizontal = "center"
			if blueColor {
				cell6R9.GetStyle().Fill.FgColor = "E8F0FE"
				cell6R9.GetStyle().Fill.PatternType = "solid"
			}

			cell7R9 := row9.AddCell()
			cell7R9.SetValue(mm[valJ.Address].JuryAgainst)
			cell7R9.GetStyle().Font.Name = "Arial"
			cell7R9.GetStyle().Font.Size = 10
			cell7R9.GetStyle().Font.Bold = true
			cell7R9.GetStyle().Alignment.Horizontal = "center"
			if blueColor {
				cell7R9.GetStyle().Fill.FgColor = "E8F0FE"
				cell7R9.GetStyle().Fill.PatternType = "solid"
			}

			indJury++
			countVote += sumVotes
			countFor += mm[valJ.Address].JuryFor
			countAbstained += mm[valJ.Address].JuryAbstained
			countAgainst += mm[valJ.Address].JuryAgainst

			blueColor = !blueColor
		}
	}

	nn := xlsx.NewStyle()
	nn.Fill.FgColor = "ACC9FE"
	nn.Fill.PatternType = "solid"

	row10 := sheet1.AddRow()
	cell1R10 := row10.AddCell()
	cell1R10.SetString(" ")
	cell1R10.GetStyle().Fill.FgColor = "ACC9FE"
	cell1R10.GetStyle().Fill.PatternType = "solid"

	cell2R10 := row10.AddCell()
	cell2R10.SetString("Total:")
	cell2R10.GetStyle().Fill.FgColor = "ACC9FE"
	cell2R10.GetStyle().Fill.PatternType = "solid"
	cell2R10.GetStyle().Font.Name = "Arial"
	cell2R10.GetStyle().Font.Size = 10
	cell2R10.GetStyle().Font.Bold = true
	cell2R10.GetStyle().Alignment.Horizontal = "right"

	cell3R10 := row10.AddCell()
	cell3R10.SetInt64(countVote)
	cell3R10.GetStyle().Fill.FgColor = "ACC9FE"
	cell3R10.GetStyle().Fill.PatternType = "solid"
	cell3R10.GetStyle().Font.Name = "Arial"
	cell3R10.GetStyle().Font.Size = 10
	cell3R10.GetStyle().Font.Bold = true
	cell3R10.GetStyle().Alignment.Horizontal = "center"

	cell4R10 := row10.AddCell()
	cell4R10.SetFloatWithFormat(sumReward, "#0.00")
	cell4R10.GetStyle().Fill.FgColor = "ACC9FE"
	cell4R10.GetStyle().Fill.PatternType = "solid"
	cell4R10.GetStyle().Font.Name = "Arial"
	cell4R10.GetStyle().Font.Size = 10
	cell4R10.GetStyle().Font.Bold = true
	cell4R10.GetStyle().Alignment.Horizontal = "center"

	cell5R10 := row10.AddCell()
	cell5R10.SetInt64(countFor)
	cell5R10.GetStyle().Fill.FgColor = "ACC9FE"
	cell5R10.GetStyle().Fill.PatternType = "solid"
	cell5R10.GetStyle().Font.Name = "Arial"
	cell5R10.GetStyle().Font.Size = 10
	cell5R10.GetStyle().Font.Bold = true
	cell5R10.GetStyle().Alignment.Horizontal = "center"

	cell6R10 := row10.AddCell()
	cell6R10.SetInt64(countAbstained)
	cell6R10.GetStyle().Fill.FgColor = "ACC9FE"
	cell6R10.GetStyle().Fill.PatternType = "solid"
	cell6R10.GetStyle().Font.Name = "Arial"
	cell6R10.GetStyle().Font.Size = 10
	cell6R10.GetStyle().Font.Bold = true
	cell6R10.GetStyle().Alignment.Horizontal = "center"

	cell7R10 := row10.AddCell()
	cell7R10.SetInt64(countAgainst)
	cell7R10.GetStyle().Fill.FgColor = "ACC9FE"
	cell7R10.GetStyle().Fill.PatternType = "solid"
	cell7R10.GetStyle().Font.Name = "Arial"
	cell7R10.GetStyle().Font.Size = 10
	cell7R10.GetStyle().Font.Bold = true
	cell7R10.GetStyle().Alignment.Horizontal = "center"

	addEmptyString(sheet1, 2, 0)
	addEmptyString(sheet1, 2, 0)

	sheet2, _ := wb.AddSheet("Result")
	sheet2.SetColWidth(0, 0, 70)
	sheet2.SetColWidth(1, 1, 20)

	styleRew1 := xlsx.NewStyle()
	styleRew1.Font.Name = "Arial"
	styleRew1.Font.Size = 10
	styleRew1.Alignment.Horizontal = "left"

	styleRew2 := xlsx.NewStyle()
	styleRew2.Font.Name = "Arial"
	styleRew2.Font.Size = 10
	styleRew1.Alignment.Horizontal = "center"

	for _, valCalc := range data.Contenders {
		if valCalc.Ranking == 0 {
			//|| valCalc.Reward == 0
			continue
		}

		row22 := sheet2.AddRow()
		cell1R2S2 := row22.AddCell()
		cell1R2S2.SetStyle(styleRew1)
		cell1R2S2.SetString(valCalc.Address)

		cell2R2S2 := row22.AddCell()
		cell2R2S2.SetStyle(styleRew2)
		cell2R2S2.SetString("0")
	}

	for _, valCalc := range data.Jurys {
		// if valCalc.Reward == 0 {
		// 	continue
		// }

		row23 := sheet2.AddRow()
		cell1R3S2 := row23.AddCell()
		cell1R3S2.SetStyle(styleRew1)
		cell1R3S2.SetString(valCalc.Address)

		cell2R3S2 := row23.AddCell()
		cell2R3S2.SetStyle(styleRew2)
		cell2R3S2.SetString("0")
	}

	err := wb.Save(data.TitleContext + ".xlsx")
	if err != nil {
		return errors.New("Error save file: " + err.Error())
	}

	return nil
}

func hexToString(in []byte) string {
	dst := make([]byte, hex.DecodedLen(len(in)))
	n, err := hex.Decode(dst, in)
	if err != nil {
		log.Fatal(err)
	}

	return string(dst[:n])
}

func addEmptyString(sheet *xlsx.Sheet, row, col int) {
	rowN := sheet.AddRow()
	cell1R3, _ := rowN.Sheet.Cell(row, col)
	cell1R3.String()
}
