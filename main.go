package main;

import (
    "os"
    "io/ioutil"
    "fmt"
    "strings"
    "errors"
    "encoding/json"
    "encoding/csv"
    "github.com/therecipe/qt/core"
    "github.com/therecipe/qt/gui"
    "github.com/therecipe/qt/widgets"
)

const DECK_DIR = "decks"

const (
    COLOR_BG            = "#1D3557"
    COLOR_BG2           = "#457B9D"
    COLOR_FG            = "#D7DCEA"
    COLOR_CARD_FRONT    = "#97D8B2"
    COLOR_CARD_BACK     = "#F68E5F"
)

func UNUSED(x ...interface{}) {}

// This type is used to list the decks in the load menu
type DeckInfo struct {
    fileName    string;
    title       string;
    cardCount   int;
}

type Card struct {
    Front   string; // The question
    Back    string; // The answer
}

// TODO: OOP
type Deck struct {
    Name    string; // Display name
    Cards   []Card; // The cards themselves
    From    string; // The name of the front side of cards
    To      string; // The name of the back side of cards
}

const (
    CARD_SET_LIST_COL_TITLE     = iota
    CARD_SET_LIST_COL_FILENAME  = iota
    CARD_SET_LIST_COL_CARDCOUNT = iota

    COUNT_OF_CARD_SET_LIST_COLS = iota
)

func readDeck(fileName string) (Deck, error) {
    var readBytes, err = ioutil.ReadFile(DECK_DIR+string(os.PathSeparator)+fileName)
    if err != nil {
        return Deck{}, err
    }
    var readStruct = Deck{}
    err = json.Unmarshal(readBytes, &readStruct)
    if err != nil {
        return Deck{}, err
    }
    return readStruct, nil
}

func readDeckInfo(fileName string) (DeckInfo, error) {
    var deck, err = readDeck(fileName)
    return DeckInfo{fileName, deck.Name, len(deck.Cards)}, err
}

func loadListItemDoubleClickedCallback(item *widgets.QTreeWidgetItem) {
    fmt.Printf("Opening deck: \"%s\"\n", item.Text(CARD_SET_LIST_COL_FILENAME))

    var window = widgets.NewQWidget(nil, 0)
    window.SetStyleSheet(fmt.Sprintf("background-color: %s; color: %s", COLOR_BG, COLOR_FG))
    window.SetWindowTitle("MemCards: "+item.Text(CARD_SET_LIST_COL_TITLE))
    window.SetFixedSize2(500, 500)

    var titleLabel = widgets.NewQLabel2(item.Text(CARD_SET_LIST_COL_TITLE), window, 0)
    titleLabel.SetAlignment(core.Qt__AlignCenter)
    titleLabel.SetFixedWidth(500)
    titleLabel.SetStyleSheet("font: 18pt")

    var deck, err = readDeck(item.Text(CARD_SET_LIST_COL_FILENAME))
    if err != nil {
        fmt.Println("Failed to load deck: \""+item.Text(CARD_SET_LIST_COL_FILENAME)+"\": "+err.Error())
        panic(err)
    }

    var activeCardI = 0
    var isActiveCardFrontSide = true
    var cardWFontSize = 18

    var cardWidget = widgets.NewQLabel2("", window, 0)
    cardWidget.SetGeometry2(20, 40, 460, 440)
    cardWidget.SetAlignment(core.Qt__AlignCenter)
    cardWidget.SetWordWrap(true)

    var cardIWidget = widgets.NewQLabel2("", window, 0)
    cardIWidget.SetGeometry2(20, window.Height()-20, window.Width()-40, 20)
    cardIWidget.SetAlignment(core.Qt__AlignCenter)

    // TODO: Restarting deck
    // TODO: Shuffling cards
    // TODO: Going back to main menu
    // TODO: Starting deck with another side visible initially (Flip all cards)

    var displayActiveCard = func() {
        var bgColor string;
        if isActiveCardFrontSide {
            bgColor = COLOR_CARD_FRONT
            cardWidget.SetText(deck.Cards[activeCardI].Front)
            cardWidget.SetToolTip(deck.From)
        } else {
            bgColor = COLOR_CARD_BACK
            cardWidget.SetText(deck.Cards[activeCardI].Back)
            cardWidget.SetToolTip(deck.To)
        }
        cardWidget.SetStyleSheet(fmt.Sprintf("background-color: %s; color: white; font: %dpt", bgColor, cardWFontSize))
        cardIWidget.SetText(fmt.Sprintf("%d/%d", activeCardI+1, len(deck.Cards)))
    }

    cardWidget.ConnectMousePressEvent(func(event *gui.QMouseEvent) {
        if event.Button() != core.Qt__LeftButton {
            return
        }

        isActiveCardFrontSide = !isActiveCardFrontSide
        displayActiveCard()
    })

    cardWidget.ConnectWheelEvent(func(event *gui.QWheelEvent){
        if event.AngleDelta().Y() > 0 {
            cardWFontSize += 2
        } else {
            cardWFontSize -= 2
        }
        if cardWFontSize < 8 {
            cardWFontSize = 8
        }
        displayActiveCard()
    })

    // Note: Hack to add shortcut to a label
    var flipButton1 = widgets.NewQPushButton(window)
    flipButton1.SetGeometry2(-100, -100, 0, 0) // Hide
    flipButton1.SetShortcut(gui.NewQKeySequence2("Up", gui.QKeySequence__NativeText))
    flipButton1.ConnectPressed(func() {
        isActiveCardFrontSide = !isActiveCardFrontSide
        displayActiveCard()
    })

    // Note: Hack to add shortcut to a label
    var flipButton2 = widgets.NewQPushButton(window)
    flipButton2.SetGeometry2(-100, -100, 0, 0) // Hide
    flipButton2.SetShortcut(gui.NewQKeySequence2("Down", gui.QKeySequence__NativeText))
    flipButton2.ConnectPressed(func() {
        isActiveCardFrontSide = !isActiveCardFrontSide
        displayActiveCard()
    })

    var goToNextCard = func() {
        activeCardI++
        if activeCardI >= len(deck.Cards)-1 {
            activeCardI = len(deck.Cards)-1
        }
        isActiveCardFrontSide = true // Flip back the cards
        displayActiveCard()
    }

    var goToPrevCard = func() {
        activeCardI--
        if activeCardI < 0 {
            activeCardI = 0
        }
        isActiveCardFrontSide = true // Flip back the cards
        displayActiveCard()
    }

    var goToPrevCardButton = widgets.NewQPushButton2("<", window)
    goToPrevCardButton.ConnectPressed(goToPrevCard)
    goToPrevCardButton.SetGeometry2(0, 40, 20, 440)
    goToPrevCardButton.SetToolTip("Go to previous card")
    goToPrevCardButton.SetShortcut(gui.NewQKeySequence2("Left", gui.QKeySequence__NativeText))

    var goToNextCardButton = widgets.NewQPushButton2(">", window)
    goToNextCardButton.SetToolTip("Go to next card")
    goToNextCardButton.ConnectPressed(goToNextCard)
    goToNextCardButton.SetGeometry2(480, 40, 20, 440)
    goToNextCardButton.SetShortcut(gui.NewQKeySequence2("Right", gui.QKeySequence__NativeText))

    displayActiveCard()

    window.Show()
}

func showLoadWinButtonCb() {
    var window = widgets.NewQWidget(nil, 0)
    window.SetStyleSheet(fmt.Sprintf("background-color: %s; color: %s", COLOR_BG, COLOR_FG))
    window.SetWindowTitle("MemCards - Load")
    window.SetFixedSize2(800, 500)

    var dirEntry, err = ioutil.ReadDir(DECK_DIR)
    if err != nil {
        var errLabel = widgets.NewQLabel2("Failed to read directory: \""+DECK_DIR+"\": "+err.Error(),
                window, 0)
        errLabel.SetGeometry2(0, 0, window.Width(), window.Height())
        errLabel.SetAlignment(core.Qt__AlignCenter)
    } else {
        var fileList []string
        for _, f := range dirEntry {
            fileList = append(fileList, f.Name())
        }
        fmt.Println("Deck files inside \""+DECK_DIR+"\":", fileList)

        var listWidget = widgets.NewQTreeWidget(window)
        listWidget.SetStyleSheet(fmt.Sprintf("selection-background-color: %s;", COLOR_BG2));
        listWidget.SetRootIsDecorated(false)
        listWidget.SetHeaderLabels([]string{"Title", "File Name", "# of Cards"})
        listWidget.SetColumnCount(COUNT_OF_CARD_SET_LIST_COLS)
        listWidget.SetGeometry2(0, 0, window.Width(), window.Height())
        listWidget.SetAllColumnsShowFocus(true)
        listWidget.ConnectItemDoubleClicked(func(item *widgets.QTreeWidgetItem, _ int) {
            window.Close()
            loadListItemDoubleClickedCallback(item)
        })

        // Note: Hack to add shortcut to a QTreeWidget
        var openButton = widgets.NewQPushButton(window)
        openButton.SetGeometry2(-100, -100, 0, 0) // Hide
        openButton.SetShortcut(gui.NewQKeySequence2("Return", gui.QKeySequence__NativeText))
        openButton.ConnectPressed(func() {
            window.Close()
            loadListItemDoubleClickedCallback(listWidget.CurrentItem())
        })

        for _, f := range fileList {
            var info, err = readDeckInfo(f)
            if err != nil {
                fmt.Printf("%s: ERROR: %s\n", f, err.Error())
            } else {
                fmt.Printf("%s: %s\n", f, info.title)
                var item = widgets.NewQTreeWidgetItem(0)
                item.SetText(CARD_SET_LIST_COL_TITLE, info.title)
                item.SetText(CARD_SET_LIST_COL_FILENAME, info.fileName)
                item.SetText(CARD_SET_LIST_COL_CARDCOUNT, fmt.Sprint(info.cardCount))
                listWidget.AddTopLevelItem(item)
            }
        }
    }

    window.Show()
}

func writeDeckToFile(filename string, deck *Deck) error {
    jsonDeck, err := json.Marshal(*deck)
    if err != nil {
        return errors.New("Error creating output JSON: "+err.Error())
    }

    fmt.Println("Writing JSON to file: "+string(jsonDeck))
    // TODO: Don't overwrite existing file
    err = ioutil.WriteFile(DECK_DIR+string(os.PathSeparator)+filename, jsonDeck, 0o644)
    if err != nil {
        return errors.New("Error writing to file: "+filename+": "+err.Error())
    }
    return nil
}

func createButtonCb(deckTitle string, deckCSV string) {
    var filename = strings.ReplaceAll(deckTitle, " ", "_")+".json"

    var reader = csv.NewReader(strings.NewReader(deckCSV))
    var cardVals, err = reader.ReadAll()
    if err != nil {
        var msgBox = widgets.NewQMessageBox2(widgets.QMessageBox__Critical, "Error",
            "Error creating deck: "+err.Error(),
            widgets.QMessageBox__Ok, nil, core.Qt__Dialog)
        msgBox.Show()
        return
    } else if len(cardVals) < 2 {
        var msgBox = widgets.NewQMessageBox2(widgets.QMessageBox__Critical, "Error",
            "Error creating deck: Not enough values in CSV.",
            widgets.QMessageBox__Ok, nil, core.Qt__Dialog)
        msgBox.Show()
        return
    }

    var deck = Deck{}
    deck.Name = strings.TrimSpace(deckTitle)
    deck.From = strings.TrimSpace(cardVals[0][0])
    deck.To = strings.TrimSpace(cardVals[0][1])
    for i, val := range cardVals {
        if i == 0 { continue }
        // TODO: Handle when lines have more/less columns than 2
        deck.Cards = append(deck.Cards, Card{Front: strings.TrimSpace(val[0]), Back: strings.TrimSpace(val[1])})
    }

    writeDeckToFile(filename, &deck)
    if err != nil {
        var msgBox = widgets.NewQMessageBox2(widgets.QMessageBox__Critical, "Error",
            err.Error(),
            widgets.QMessageBox__Ok, nil, core.Qt__Dialog)
        msgBox.Show()
        return
    }

    var msgBox = widgets.NewQMessageBox2(widgets.QMessageBox__Information, "Deck Created",
        fmt.Sprintf("Created a deck.\nTitle: %s\nFilename: %s\n# of cards: %d", deck.Name, filename, len(deck.Cards)),
        widgets.QMessageBox__Ok, nil, core.Qt__Dialog)
    msgBox.Show()

    // TODO: Go back to main menu after creating deck
}

func showCreateWinButtonCb() {
    var window = widgets.NewQWidget(nil, 0)
    window.SetStyleSheet(fmt.Sprintf("background-color: %s; color: %s", COLOR_BG, COLOR_FG))
    window.SetWindowTitle("MemCards - Create - \"Unnamed\"")
    window.SetFixedSize2(800, 500)

    var titleEntry = widgets.NewQLineEdit2("Unnamed", window)
    titleEntry.SetGeometry2(0, 0, 800, 30)
    titleEntry.SetStyleSheet("font: 18pt")
    titleEntry.SetAlignment(core.Qt__AlignCenter)
    titleEntry.ConnectTextEdited(func(val string) {
        window.SetWindowTitle("MemCards - Create - \""+val+"\"")
    })

    var textWidget = widgets.NewQPlainTextEdit(window)
    textWidget.SetGeometry2(0, 30, window.Width(), window.Height()-50)
    textWidget.SetToolTip("Enter the card values here in CSV format.\n"+
        "Left column is the front, right column is the back side of cards."+
        "The first line specifies the name of sides.")

    var createButton = widgets.NewQPushButton2("Create", window)
    createButton.SetGeometry2(0, window.Height()-20, window.Width(), 20);
    createButton.ConnectPressed(func() {
        window.Close()
        createButtonCb(strings.TrimSpace(titleEntry.Text()), textWidget.ToPlainText())
    })

    window.Show()
}

// TODO: Editing decks
// TODO: Catgorizing decks (subfolders?)
// TODO: Coloring decks (by category?)

func main() {
    var app = gui.NewQGuiApplication(len(os.Args), os.Args)

    var window = widgets.NewQWidget(nil, 0)
    window.SetStyleSheet(fmt.Sprintf("background-color: %s; color: %s", COLOR_BG, COLOR_FG))
    window.SetWindowTitle("MemCards")
    window.SetFixedSize2(500, 200)

    var loadButton = widgets.NewQPushButton2("&Load", window)
    loadButton.SetShortcut(gui.NewQKeySequence2("L", gui.QKeySequence__NativeText))
    loadButton.SetStyleSheet("font: 18pt")
    loadButton.SetGeometry2(0, 0, 500, 100)
    loadButton.ConnectPressed(func() {
        window.Close()
        showLoadWinButtonCb()
    })

    var createButton = widgets.NewQPushButton2("&Create", window)
    createButton.SetShortcut(gui.NewQKeySequence2("C", gui.QKeySequence__NativeText))
    createButton.SetStyleSheet("font: 18pt")
    createButton.SetGeometry2(0, 100, 500, 100)
    createButton.ConnectPressed(func() {
        window.Close()
        showCreateWinButtonCb()
    })

    window.Show()
    app.Exec()
}
