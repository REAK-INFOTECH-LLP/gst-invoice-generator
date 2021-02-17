package main

import (
    "text/template"
    "encoding/json"
    "io/ioutil"
    "os"
    "fmt"
    "strings"
    "math"
    "regexp"
    "os/exec"
    "github.com/divan/num2words"
)

type InvoiceItems struct {
    Description string
    Hsn int32
    Qty float32
    RatePerItem float32
    Total float32
    DiscountPercentage float32
    TaxableValue float32
    IgstAmount float32
    CgstAmount float32
    SgstAmount float32
}

type Invoice struct{
    InvoiceNumber string
    InvoiceDate string
    InvoiceAmount int32
    DueDate string
    Items []InvoiceItems
}

type InvoiceFrom struct {
    Name string
    GstNumber string
    AddressLine1 string
    AddressLine2 string
    AddressLine3 string
    AddressLine4 string
}

type InvoiceToInfo struct {
    Name string
    AddressLine1 string
    AddressLine2 string
    AddressLine3 string
    AddressLine4 string
}

type InvoiceTo struct {
    GstNumber string
    Billing InvoiceToInfo
    Shipping InvoiceToInfo
}

type Settings struct {
    Logo string
    PaymentTerms string
    ImagePath string
    IgstRate float32
    CgstRate float32
    SgstRate float32
}

type Data struct {
    Invoice Invoice
    InvoiceFrom InvoiceFrom
    InvoiceTo InvoiceTo
    Settings Settings
    Auto Auto
}

type Auto struct {
    DueDate string
    StateOfSupply string
    GstType string
    TotalNumeral float32
    TotalWords string
    ChargedGST float32
}

func main(){
    // Load Data JSON
    jsonFile, err := os.Open("data.json") 
    if err != nil {
        fmt.Println(err)
    }
    byteValue, _ := ioutil.ReadAll(jsonFile)
    defer jsonFile.Close()

    var data Data
    json.Unmarshal(byteValue, &data)

    // Compute Auto Fields
    logic(&data)
    
    t, err := template.ParseFiles("invoice.tex")
    //fmt.Println(t, err)

    var replacer = strings.NewReplacer(" ", "_", "/", "_")
    rx, _ := regexp.Compile("([0-9/])+")
    outputFilename := rx.FindString(data.Invoice.InvoiceNumber)
    outputFilename = replacer.Replace(outputFilename)

    f, err := os.Create("temp/tmp.tex")
    if err != nil {
        fmt.Println("create file: ", err)
        return
    }

    err = t.Execute(f, data)
    if err != nil {
        fmt.Print("execute: ", err)
        return
    }

    exec.Command("pdflatex","temp/tmp.tex").Output()
    exec.Command("mv", "tmp.pdf", "output/"+outputFilename+".pdf").Output()
    exec.Command("rm", "-rf", "tmp*").Output()
}

func logic(data *Data) {
    calculateLineItem(data)
}

func calculateLineItem(data *Data){
    var runningTotal float32
    var taxTotal float32
    for index, item := range data.Invoice.Items {
        data.Invoice.Items[index].Total = item.Qty * item.RatePerItem

        data.Invoice.Items[index].TaxableValue = data.Invoice.Items[index].Total - (data.Invoice.Items[index].Total * (item.DiscountPercentage / 100))
      
        if(data.Auto.GstType == "IGST"){
            data.Invoice.Items[index].IgstAmount = data.Invoice.Items[index].TaxableValue * (data.Settings.IgstRate / 100)
            data.Invoice.Items[index].IgstAmount = float32(math.Round((float64(data.Invoice.Items[index].IgstAmount)*100)/100))
            taxTotal += data.Invoice.Items[index].IgstAmount
        }
        
        if(data.Auto.GstType == "CGST"){
            data.Invoice.Items[index].CgstAmount = data.Invoice.Items[index].TaxableValue * (data.Settings.CgstRate / 100)
            data.Invoice.Items[index].CgstAmount = float32(math.Round((float64(data.Invoice.Items[index].CgstAmount)*100)/100))
            data.Invoice.Items[index].SgstAmount = data.Invoice.Items[index].TaxableValue * (data.Settings.SgstRate / 100)
            data.Invoice.Items[index].SgstAmount = float32(math.Round((float64(data.Invoice.Items[index].SgstAmount)*100)/100))
            taxTotal += data.Invoice.Items[index].CgstAmount + data.Invoice.Items[index].SgstAmount
        }


        runningTotal += float32(math.Round(float64(data.Invoice.Items[index].TaxableValue)*100)/100)
    }

    data.Auto.TotalNumeral = runningTotal
    data.Auto.TotalWords = strings.Title(num2words.ConvertAnd(int(runningTotal)))
    data.Auto.ChargedGST = taxTotal
}
