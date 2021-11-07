package main;

import (
    "os"
    "io/ioutil"
    "fmt"
    "encoding/json"
    "github.com/therecipe/qt/core"
    "github.com/therecipe/qt/gui"
    "github.com/therecipe/qt/widgets"
)

const CARD_SET_DIR = "cardsets"

func UNUSED(x ...interface{}) {}

// This type is used to list the card sets in the load menu
type CardSetInfo struct {
    fileName    string;
    title       string;
    cardCount   int;
}

type Card struct {
    Front   string; // The question
    Back    string; // The answer
}

type CardSet struct {
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

func readCardSet(fileName string) (CardSet, error) {
    var readBytes, err = ioutil.ReadFile(CARD_SET_DIR+string(os.PathSeparator)+fileName)
    if err != nil {
        return CardSet{}, err
    }
    var readStruct = CardSet{}
    err = json.Unmarshal(readBytes, &readStruct)
    if err != nil {
        return CardSet{}, err
    }
    return readStruct, nil
}

func readCardsetInfo(fileName string) (CardSetInfo, error) {
    var cardSet, err = readCardSet(fileName)
    return CardSetInfo{fileName, cardSet.Name, len(cardSet.Cards)}, err
}

func loadListItemDoubleClickedCallback(item *widgets.QTreeWidgetItem) {
    fmt.Printf("Opening card set: \"%s\"\n", item.Text(CARD_SET_LIST_COL_FILENAME))

    var window = widgets.NewQWidget(nil, 0)
    window.SetWindowTitle("MemCards: "+item.Text(CARD_SET_LIST_COL_TITLE))
    window.SetFixedSize2(500, 500)

    var titleLabel = widgets.NewQLabel2(item.Text(CARD_SET_LIST_COL_TITLE), window, 0)
    titleLabel.SetAlignment(core.Qt__AlignCenter)
    titleLabel.SetFixedWidth(500)
    titleLabel.SetStyleSheet("font: 18pt")

    var cardSet, err = readCardSet(item.Text(CARD_SET_LIST_COL_FILENAME))
    if err != nil {
        fmt.Println("Failed to load card set: \""+item.Text(CARD_SET_LIST_COL_FILENAME)+"\": "+err.Error())
        panic(err)
    }

    var activeCardI = 0
    var isActiveCardFrontSide = true

    var cardWidget = widgets.NewQLabel2("", window, 0)
    cardWidget.SetGeometry2(20, 40, 460, 440)
    cardWidget.SetAlignment(core.Qt__AlignCenter)
    cardWidget.SetWordWrap(true)

    var displayActiveCard = func() {
        var bgColor string;
        if isActiveCardFrontSide {
            bgColor = "#293CCA"
            cardWidget.SetText(cardSet.Cards[activeCardI].Front)
            cardWidget.SetToolTip(cardSet.From)
        } else {
            bgColor = "#C8851F"
            cardWidget.SetText(cardSet.Cards[activeCardI].Back)
            cardWidget.SetToolTip(cardSet.To)
        }
        cardWidget.SetStyleSheet(fmt.Sprintf("background-color: %s; color: white", bgColor))
    }

    cardWidget.ConnectMousePressEvent(func(event *gui.QMouseEvent) {
        if event.Button() != core.Qt__LeftButton {
            return
        }

        isActiveCardFrontSide = !isActiveCardFrontSide
        displayActiveCard()
    })

    // Note: Hack to add shortcut to a label
    var flipButton = widgets.NewQPushButton(window)
    flipButton.SetGeometry2(-100, -100, 0, 0) // Hide
    flipButton.SetShortcut(gui.NewQKeySequence2("Up", gui.QKeySequence__NativeText))
    flipButton.ConnectPressed(func() {
        isActiveCardFrontSide = !isActiveCardFrontSide
        displayActiveCard()
    })

    var goToNextCard = func() {
        activeCardI++
        if activeCardI >= len(cardSet.Cards)-1 {
            activeCardI = len(cardSet.Cards)-1
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

func loadButtonCallback() {
    var window = widgets.NewQWidget(nil, 0)
    window.SetWindowTitle("MemCards - Load")
    window.SetFixedSize2(800, 500)

    var dirEntry, err = ioutil.ReadDir(CARD_SET_DIR)
    if err != nil {
        var errLabel = widgets.NewQLabel2("Failed to read directory: \""+CARD_SET_DIR+"\": "+err.Error(),
                window, 0)
        errLabel.SetGeometry2(0, 0, window.Width(), window.Height())
        errLabel.SetAlignment(core.Qt__AlignCenter)
    } else {
        var fileList []string
        for _, f := range dirEntry {
            fileList = append(fileList, f.Name())
        }
        fmt.Println("Card set files inside \""+CARD_SET_DIR+"\":", fileList)

        var listWidget = widgets.NewQTreeWidget(window)
        listWidget.SetRootIsDecorated(false)
        listWidget.SetHeaderLabels([]string{"Title", "File Name", "# of Cards"})
        listWidget.SetColumnCount(COUNT_OF_CARD_SET_LIST_COLS)
        listWidget.SetGeometry2(0, 0, window.Width(), window.Height())
        listWidget.SetAllColumnsShowFocus(true)
        listWidget.SetAlternatingRowColors(true)
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
            var info, err = readCardsetInfo(f)
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

func main() {
    var app = gui.NewQGuiApplication(len(os.Args), os.Args)

    var window = widgets.NewQWidget(nil, 0)
    window.SetWindowTitle("MemCards")
    window.SetFixedSize2(500, 200)

    var loadButton = widgets.NewQPushButton2("&Load", window)
    loadButton.SetShortcut(gui.NewQKeySequence2("L", gui.QKeySequence__NativeText))
    loadButton.SetStyleSheet("font: 18pt")
    loadButton.SetGeometry2(0, 0, 500, 100)
    loadButton.ConnectPressed(func() {
        window.Close()
        loadButtonCallback()
    })

    var createButton = widgets.NewQPushButton2("&Create", window)
    createButton.SetShortcut(gui.NewQKeySequence2("C", gui.QKeySequence__NativeText))
    createButton.SetStyleSheet("font: 18pt")
    createButton.SetGeometry2(0, 100, 500, 100)

    window.Show()
    app.Exec()
}
