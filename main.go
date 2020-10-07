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

	goton "github.com/move-ton/ton-client-go"
	contracts "github.com/move-ton/ton-client-go/contracts"
	"github.com/tealeg/xlsx/v3"
)

var (
	networkChainID int
	addressContest string
)

func init() {
	flag.IntVar(&networkChainID, "chain", 0, "0-mainnet, 1-devnet")
	flag.StringVar(&addressContest, "addr", "", "Address contract contest")
}

//addressContest := "0:3a139813e5fd8427ec0700b85de063a0a574f54b8a942adcd55d5b60e72aa76b" //wiki
// addressContest := "0:e06975f462608516a891c1a62704f75ad4bd71df02c39cc79259bd623f9da148" // DEFI

func main() {

	flag.Parse()

	nameChain := "main.ton.dev"
	if addressContest == "" {
		log.Fatal("Parameter addr is empty!")
	}

	if networkChainID != 0 {
		nameChain = "net.ton.dev"
	}

	config, err := goton.ParseConfigFile("config.toml")
	if err != nil {
		log.Println("Error read config file, err: . Settings setup on default.", err)
		config = goton.NewConfig(0)
	}

	config.Servers[0] = nameChain

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

	abiS := &contracts.ABI{}
	err = json.Unmarshal(abiByte, abiS)

	pOLR := &contracts.ParamsOfLocalRun{}
	pOLR.Abi = *abiS
	pOLR.Address = addressContest
	pOLR.FunctionName = "listContenders"
	pOLR.Input = T{}

	md := &mainDats{}

	result, err := contracts.RunLocalResp(client.Request(contracts.RunLocal(pOLR)))
	if err != nil {
		fmt.Println("Error result: ", err)
		return
	}

	res1 := &resultContenders{}
	_ = json.Unmarshal(*result.Output, res1)

	lenC := len(res1.Addresses)
	for i := 0; i < lenC; i++ {
		contDrs := &contenders{}
		contDrs.IDS = res1.Ids[i]
		contDrs.Address = res1.Addresses[i]

		md.Contenders = append(md.Contenders, contDrs)
	}

	pOLR.FunctionName = "getContestInfo"
	result, err = contracts.RunLocalResp(client.Request(contracts.RunLocal(pOLR)))
	if err != nil {
		fmt.Println("Error result: ", err)
		return
	}

	res2 := &resContestInfo{}
	_ = json.Unmarshal(*result.Output, res2)

	md.TitleContext = hexToString([]byte(res2.Title))
	md.LinkToContext = hexToString([]byte(res2.Link))
	lenJ := len(res2.JuryKeys)
	for i := 0; i < lenJ; i++ {
		juryS := &jury{}
		juryS.Address = res2.JuryAddresses[i]
		juryS.PublicKey = res2.JuryKeys[i]

		md.Jurys = append(md.Jurys, juryS)
	}

	mm := make(map[string]votes)
	for n, val := range md.Contenders {
		id, err := strconv.ParseInt(val.IDS, 0, 16)
		if err != nil {
			log.Fatalln("don't parse int from string id to int16: ", err)
		}
		pOLR.FunctionName = "getVotesPerJuror"

		idReq := req{ID: id}
		pOLR.Input = idReq
		result, err = contracts.RunLocalResp(client.Request(contracts.RunLocal(pOLR)))
		if err != nil {
			fmt.Println("Error result: ", err)
			return
		}

		res3 := &goverment{}
		_ = json.Unmarshal(*result.Output, res3)

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

	sort.Slice(md.Contenders, func(i, j int) bool {
		return md.Contenders[i].AverageScore > md.Contenders[j].AverageScore
	})

	var (
		intprev float64
		place   int64
	)

	countJury := int64(len(md.Jurys))
	place = 1
	for _, valueC := range md.Contenders {

		if math.IsNaN(valueC.AverageScore) || valueC.AverageScore == 0 || (valueC.Reject >= (countJury)+1) {
			continue
		}

		if place == 1 {
			intprev = valueC.AverageScore
			valueC.Ranking = place
			place++
		} else {
			if valueC.AverageScore < intprev {
				intprev = valueC.AverageScore
				valueC.Ranking = place
				place++
			} else {
				valueC.Ranking = place - 1
			}
		}
	}

	sort.Slice(md.Contenders, func(i, j int) bool {
		return md.Contenders[i].IDS < md.Contenders[j].IDS
	})

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

	cell5R5 := row4.AddCell()
	cell5R5.SetString("Reject")
	cell5R5.SetStyle(st)

	blueColor := false
	for _, val := range data.Contenders {
		id, _ := strconv.ParseInt(val.IDS, 0, 16)
		row5 := sheet1.AddRow()
		cell1R5 := row5.AddCell()
		cell1R5.SetValue(id)
		cell1R5.GetStyle().Font.Name = "Arial"
		cell1R5.GetStyle().Font.Size = 10
		cell1R5.GetStyle().Alignment.Horizontal = "center"
		cell1R5.GetStyle().Font.Bold = true
		if blueColor {
			cell1R5.GetStyle().Fill.FgColor = "E8F0FE"
			cell1R5.GetStyle().Fill.PatternType = "solid"
		}

		cell2R5 := row5.AddCell()
		cell2R5.SetHyperlink(linkToExplorer+val.Address, val.Address, "")
		cell2R5.GetStyle().Font.Size = 10
		cell2R5.GetStyle().Font.Color = "1155CC"
		cell2R5.GetStyle().Font.Underline = true
		if blueColor {
			cell2R5.GetStyle().Fill.FgColor = "E8F0FE"
			cell2R5.GetStyle().Fill.PatternType = "solid"
		}

		cell3R5 := row5.AddCell()
		cell3R5.SetFloatWithFormat(val.AverageScore, "#0.00")
		cell3R5.GetStyle().Font.Name = "Arial"
		cell3R5.GetStyle().Font.Size = 10
		cell3R5.GetStyle().Alignment.Horizontal = "center"
		cell3R5.GetStyle().Font.Bold = true
		if blueColor {
			cell3R5.GetStyle().Fill.FgColor = "E8F0FE"
			cell3R5.GetStyle().Fill.PatternType = "solid"
		}

		cell4R5 := row5.AddCell()
		cell4R5.GetStyle().Font.Name = "Arial"
		cell4R5.GetStyle().Font.Size = 10
		cell4R5.GetStyle().Font.Bold = true
		cell4R5.SetInt64(val.Ranking)
		cell4R5.GetStyle().Alignment.Horizontal = "center"
		if blueColor {
			cell4R5.GetStyle().Fill.FgColor = "E8F0FE"
			cell4R5.GetStyle().Fill.PatternType = "solid"
		}

		cell5R9 := row5.AddCell()
		cell5R9.GetStyle().Font.Name = "Arial"
		cell5R9.GetStyle().Font.Size = 10
		cell5R9.GetStyle().Font.Bold = true
		cell5R9.SetFloatWithFormat(0, "#0.00")
		cell5R9.GetStyle().Alignment.Horizontal = "center"
		if blueColor {
			cell5R9.GetStyle().Fill.FgColor = "E8F0FE"
			cell5R9.GetStyle().Fill.PatternType = "solid"
		}

		cell6R9 := row5.AddCell()
		cell6R9.SetInt64(val.Reject)
		cell6R9.GetStyle().Font.Name = "Arial"
		cell6R9.GetStyle().Font.Size = 10
		cell6R9.GetStyle().Alignment.Horizontal = "center"
		cell6R9.GetStyle().Font.Bold = true
		if blueColor {
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
	cell5R45.SetString(" ")
	cell5R45.SetStyle(stylefoo1)

	cell6R45 := row45.AddCell()
	cell6R45.SetString(" ")
	cell6R45.SetStyle(stylefoo1)

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
	cell5R8.SetString("For")
	cell5R8.SetStyle(st)

	cell6R8 := row8.AddCell()
	cell6R8.SetString("Abstained")
	cell6R8.SetStyle(st)

	cell7R8 := row8.AddCell()
	cell7R8.SetString("Against")
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
	sheet2.SetColWidth(1, 1, 70)
	sheet2.SetColWidth(2, 6, 20)

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
