import {registerLocale as register} from 'amis-core';

// =========================================================
// Af Soomaali (so-SO) - Luuqadda UI ee amis-core
//
// Qoraaga: AwoERP / Anika Global Limited
// Taariikhda: 2026
//
// Xusuusin ku saabsan tarjumaadda:
// - Ereyada teknoolajiyada ee aan lahayn beddelka Soomaali
//   ayaa loo adeegsaday ereygii ugu dhowaa macnaheeda.
// - Xaaladaha ay jiraan labada doorasho oo labaduba
//   sax yihiin, waxaa la doortay eraygii ugu caansan.
// - Xarfaha ASCII sida {{variable}} waa la dhaafsaday
//   siday yihiin maadaama ay koodh yihiin, ma tarjumna.
// =========================================================

register('so-SO', {

  // -------------------------------------------------------
  // FICILADA (Actions)
  // "Wait for Xs" → "Sugta {{timeLeft}} ilbiriqsi"
  // "ilbiriqsi" = second (unit of time in Somali)
  // -------------------------------------------------------
  'Action.countDown': 'Sugta ${timeLeft} ilbiriqsi',

  // -------------------------------------------------------
  // OGEYSIISKA (Alert)
  // "System Info" → "Macluumaadka Nidaamka"
  // -------------------------------------------------------
  'Alert.info': 'Macluumaadka Nidaamka',

  // -------------------------------------------------------
  // CODSIGA (App navigation)
  // "Home" → "Guriga" (lit. home/house, used universally for home screen)
  // "Navigation" → "Xulashada Meelaha" (lit. selecting places/navigation)
  // -------------------------------------------------------
  'App.home': 'Guriga',
  'App.navigation': 'Xulashada Meelaha',

  // -------------------------------------------------------
  // KALA SOOCIDA (Sorting)
  // "Asc" → "Kor u kaca" (ascending = going upward)
  // "Desc" → "Hoos u dhaca" (descending = going downward)
  // -------------------------------------------------------
  'asc': 'Kor u kaca',
  'desc': 'Hoos u dhaca',

  // -------------------------------------------------------
  // KALANDARKA (Calendar)
  // "Date Picker" → "Dooro Taariikhda"
  // "begin/end" waa xeeladaha gaagaaban ee loogu talagalay
  // bannaanka yar ee muujinta bilawga iyo dhamaadka
  // -------------------------------------------------------
  'Calendar.datepicker': 'Dooro Taariikhda',
  'Calendar.yearmonth': 'SSSB', // Sannad Shaharkiisa Bilowga - keeping YYYY MM pattern concept
  'Calendar.year': 'SSSS',      // Sannad - YYYY equivalent
  'Calendar.begin': 'bilow',
  'Calendar.end': 'dhamaad',
  'Calendar.beginAndEnd': 'b/d',  // bilow/dhamaad (abbreviated)
  'Calendar.toast': 'Ka baxsan xadka taariikhda',
  'Calendar.startPick': 'Dooro wakhtiga bilowga',
  'Calendar.endPick': 'Dooro wakhtiga dhamaadka',

  // -------------------------------------------------------
  // FICILADA GUUD (General actions)
  // -------------------------------------------------------
  'cancel': 'Jooji',       // Cancel = stop/abort an action
  'confirm': 'Xaqiiji',   // Confirm = verify/affirm
  'clear': 'Nadiifi',     // Clear = clean/wipe
  'more': 'Wax dheeraad ah',
  'delete': 'Tirtir',
  'reset': 'Dib u deji',
  'save': 'Keydi',
  'search': 'Raadi',
  'sort': 'Kala sooc',
  'stop': 'Jooji',
  'reload': 'Dib u soo geli',
  'loading': 'Waa la rarayo...',
  'link': 'Xiriir',
  'add': 'Ku dar',         // "New" in context of adding items

  // -------------------------------------------------------
  // KAARARKA (Card)
  // -------------------------------------------------------
  'Card.dragTip': 'Jiid badhanka kore si aad u kala soocdo',
  'Card.toggleDrag': 'Bedel hab-jiidashada kala soocida',

  // -------------------------------------------------------
  // MAGAALADA (City)
  // "street" → "waddada" (road/street)
  // -------------------------------------------------------
  'City.street': 'Geli macluumaadka waddada',

  // -------------------------------------------------------
  // MIDABKA (Color Picker)
  // -------------------------------------------------------
  'ColorPicker.placeholder': 'Dooro midab',

  // -------------------------------------------------------
  // QAABKA HOOSAADKA (SubForm / Combo)
  // "SubForm.add" and "add" are both "New" in English
  // using "Ku dar" (add) for consistency
  // -------------------------------------------------------
  'SubForm.add': 'Ku dar',
  'SubForm.button': 'Habayn',       // "Config" → configure/setup
  'SubForm.editDetail': 'Wax ka beddel faahfaahinta',

  // -------------------------------------------------------
  // COMBO (editable list/array field)
  // -------------------------------------------------------
  'Combo.dragDropSort': 'Jiid si aad u kala soocdo',
  'Combo.invalidData': 'Xogta waa khaldan tahay, fadlan tirtir',
  'Combo.maxLength': 'Tiradda ugu badan waa {{MaxLength}}. Fadlan tirtir qaar',
  'Combo.minLength': 'Ugu yaraan {{minLength}} shay ayaa loo baahan yahay. Fadlan ku dar',
  'Combo.type': 'Nooca',

  // -------------------------------------------------------
  // KOOBIYEYNTA (Copyable)
  // -------------------------------------------------------
  'Copyable.tip': 'Koobiyee',

  // -------------------------------------------------------
  // CRUD (List/Table data management)
  // CRUD = Create, Read, Update, Delete
  // "Export" → "Dhoofin" (to export/send out)
  // -------------------------------------------------------
  'CRUD.exportCSV': 'Dhoofin CSV',
  'CRUD.exportExcel': 'Dhoofin Excel',
  'CRUD.exportExcelTemplate': 'Dhoofin Qaabka Excel',
  'CRUD.fetchFailed': 'Soo dejinta waa la guuldareystay',
  'CRUD.filter': 'Shid',            // "Filter" → sift/screen
  'CRUD.selected': 'Waxaa la doortay {{total}} shay: ',
  'CRUD.invalidArray': 'data.items waa inay ahaataa array',
  'CRUD.invalidData': 'Xogtu way maqan tahay',
  'CRUD.loadMore': 'Soo rar wax dheeraad ah',
  'CRUD.loadMoreDisableTip': 'Xog ma jirto ama waa bogga u dambeeya',
  'CRUD.perPage': 'Bogga kasta',
  'CRUD.stat': '{{page}} ka {{lastPage}} wadarta: {{total}}.',
  'CRUD.paginationGoText': 'Aad',
  'CRUD.paginationPageText': 'bog',
  'CRUD.confirmLeaveUnSavedPage':
    'Beddelka bogga waxa uu lumiyaa xogta aan la keydin, fadlan xaqiiji.',

  // -------------------------------------------------------
  // BOGGA (Pagination)
  // -------------------------------------------------------
  'PaginationWrapper.placeholder': 'Fadlan habayn jirka',
  'Pagination.select': '{{count}} shay/bog',
  'Pagination.goto': 'Aad',
  'Pagination.go': 'AADI',
  'Pagination.totalPage': 'wadarta {{lastPage}} bog',
  'Pagination.totalCount': 'wadarta {{total}} shay',

  // -------------------------------------------------------
  // TAARIIKHDA (Date)
  // "Select" → "Dooro" throughout
  // -------------------------------------------------------
  'Date.titleYear': 'dooro sannad',
  'Date.titleMonth': 'dooro bishii iyo sannadka',
  'Date.titleQuarter': 'dooro rubuc sannadeedka',
  'Date.titleDate': 'dooro bishii, sannadka iyo maalinta',
  'Date.titleTime': 'dooro wakhtiga',
  'Date.daysago': '{{days}} maalin ka hor',
  'Date.dayslater': '{{days}} maalin ka dib',
  'Date.endOfMonth': 'Maalinta u dambaysa ee bishii',
  'Date.endOfLastMonth': 'Maalinta u dambaysa ee bishii hore',
  'Date.endOfWeek': 'Sabti',
  'Date.hoursago': '{{hours}} saacadood ka hor',
  'Date.hourslater': '{{hours}} saacadood ka dib',
  'Date.invalid': 'Taariikhdu waa khaldan tahay',
  'Number.invalid': 'Lambarka waa khaldan yahay',
  'Date.monday': 'Isniin',
  'Date.monthsago': '{{months}} bil ka hor',
  'Date.monthslater': '{{months}} bil ka dib',
  'Date.now': 'Hadda',
  'Date.placeholder': 'Dooro Taariikhda',
  'Date.quartersago': '{{quarters}} rubuc sannadeed ka hor',
  'Date.quarterslater': '{{quarters}} rubuc sannadeed ka dib',
  'Date.startOfLastMonth': 'Maalinta koowaad ee bishii hore',
  'Date.startOfLastQuarter': 'Maalinta koowaad ee rubucii hore',
  'Date.startOfMonth': 'Maalinta koowaad ee bishii',
  'Date.startOfQuarter': 'Maalinta koowaad ee rubucii',
  'Date.today': 'Maanta',
  'Date.tomorrow': 'Berri',
  'Date.weeksago': '{{weeks}} toddobaad ka hor',
  'Date.weekslater': '{{weeks}} toddobaad ka dib',
  'Date.yesterday': 'Shalay',
  'dateformat.year': 'SSSS',  // Year format placeholder

  // -------------------------------------------------------
  // XADKA TAARIIKHDA (Date Range)
  // -------------------------------------------------------
  'DateRange.daysago': 'Maalmihii {{days}} ee ugu dambeeyay',
  'DateRange.dayslater': 'Gudaha {{days}} maalmood',
  'DateRange.weeksago': 'Toddobaadyadii {{weeks}} ee ugu dambeeyay',
  'DateRange.weekslater': 'Gudaha {{weeks}} toddobaad',
  'DateRange.monthsago': 'Bilihii {{months}} ee ugu dambeeyay',
  'DateRange.monthslater': 'Gudaha {{months}} bil',
  'DateRange.quartersago': 'Rubucyadii {{quarters}} ee ugu dambeeyay',
  'DateRange.quarterslater': 'Gudaha {{quarters}} rubuc',
  'DateRange.yearsago': 'Sannadihii {{years}} ee ugu dambeeyay',
  'DateRange.yearslater': 'Gudaha {{years}} sannadood',
  'DateRange.hoursago': 'Saacadihii {{hours}} ee ugu dambeeyay',
  'DateRange.hourslater': 'Gudaha {{hours}} saacadood',
  'DateRange.1dayago': '1 maalin ka hor',
  'DateRange.1daysago': '1 maalin ka hor',
  'DateRange.7daysago': '7 maalmood ka hor',
  'DateRange.30daysago': '30 maalmood ka hor',
  'DateRange.90daysago': '90 maalmood ka hor',
  'DateRange.lastMonth': 'Bishii hore',
  'DateRange.lastWeek': 'Toddobaadkii hore',
  'DateRange.lastQuarter': 'Rubucii hore',
  'DateRange.placeholder': 'Dooro xadka taariikhda',
  'YearRange.placeholder': 'Dooro xadka sannadka',
  'DateRange.thisWeek': 'Toddobaadkan',
  'DateRange.thisMonth': 'Bisha',
  'DateRange.thisQuarter': 'Rubucaan',
  'DateRange.thisYear': 'Sannadkan',
  'DateRange.lastYear': 'Sannadkii hore',
  'DateRange.valueConcat': ' ilaa ',   // "to" as in "Jan to Feb" → "ilaa" (until/to)

  // -------------------------------------------------------
  // WAKHTI-TAARIIKHDA (DateTime)
  // -------------------------------------------------------
  'DateTime.placeholder': 'Dooro Taariikhda iyo Wakhtiga',

  // -------------------------------------------------------
  // TIRTIRKA (Delete)
  // -------------------------------------------------------
  'deleteConfirm': 'Ma hubtaa inaad tirtirayso?',
  'deleteFailed': 'Tirtirku wuu guuldareystay',

  // -------------------------------------------------------
  // WADAAGGA (Dialog)
  // "Close" → "Xidh" (to close/shut)
  // -------------------------------------------------------
  'Dialog.close': 'Xidh',
  'Dialog.screen': 'Muuqaalka Buuxa',         // Fullscreen mode
  'Dialog.reset': 'Ka bax muuqaalka buuxa',
  'Dialog.title': 'Cinwaanka',

  // -------------------------------------------------------
  // KU-DHEJINTA (Embed)
  // -------------------------------------------------------
  'Embed.invalidRoot': 'Xiddiga xididdada waa khaldan tahay',
  'Embed.downloading': 'Soo dejinta ayaa bilaabmaysa',

  // -------------------------------------------------------
  // EXCEL GELINTA (Excel upload)
  // -------------------------------------------------------
  'Excel.placeholder': 'Jiid oo dhig faylka Excel halkan, ama guji si aad u doorato',
  'Excel.parsed': '{{filename}} waa la falanqeeyay',

  // -------------------------------------------------------
  // KHALADAADKA GUUD (General errors/status)
  // -------------------------------------------------------
  'fetchFailed': 'Codsiga API wuu guuldareystay',
  'loadingFailed': 'Rarista way guuldareystay',
  'saveFailed': 'Keydinta way guuldareystay',
  'saveSuccess': 'Si guul leh ayaa loo keydiday',
  'validateFailed': 'Xaqiijinta way guuldareystay',
  'networkError': 'Khalad shabakad ah ama habayn CORS ah oo maqan',
  'noResult': 'Natiijo la ma helin',
  'searchHistory': 'Taariikhda raadinta',
  'searchResult': 'Natiijada raadinta',

  // -------------------------------------------------------
  // FAYLKA (File upload)
  // "Upload" → "Kor u dhig" (lit. put up/upload)
  // -------------------------------------------------------
  'File.continueAdd': 'Sii wad ku darida',
  'File.downloadTpl': 'Soo dajiso qaabka',
  'File.dragDrop': 'Jiid oo dhig faylalka halkan ama',
  'File.clickUpload': 'guji halkan si aad u kordhiso',
  'File.helpText': 'Dukumiintiga caawinada',
  'File.errorRetry': 'Kor u dhigistii faylka way guuldareystay, fadlan isku day mar kale',
  'File.failed': 'Faylalka guuldareystay.',
  'File.invalidType': '{{files}} kuma habboona nooca `{{accept}}`',
  'File.maxLength': 'Tiradda ugu badan waa {{maxLength}}',
  'File.maxSize':
    '{{filename}} aad doortay waxay ka baxaysaa xadka ugu badan ee {{maxSize}}',
  'File.pause': 'Hakad geli kor u dhigista',
  'File.repick': 'Dib u dooro',
  'File.result':
    'Si guul leh {{uploaded}} fayl ayaa la kordhiyay, {{failed}} fayl ayaa guuldareystay',
  'File.retry': 'Isku day mar kale',
  'File.sizeLimit': 'Cabbirka faylka ugu badan waa {{maxSize}}',
  'File.start': 'Bilow kor u dhigista',
  'File.upload': 'Kor u dhig',
  'File.uploadFailed': 'Xogta laga soo celiyay API-ga kor u dhigista waa maran tahay',
  'File.uploading': 'Waa la kordhiyayaa',

  // -------------------------------------------------------
  // FOOMKA (Form)
  // "Submit" → "Dir" (to send/submit)
  // -------------------------------------------------------
  'FormItem.autoFillLoadFailed': 'Xogta laga soo celiyay API-ga cusbooneysiga tooska ah waa khaldan tahay',
  'FormItem.autoFillSuggest': 'Tixraac xogta',
  'Form.loadOptionsFailed': 'Doorashooyinka soo dejista way guuldareystay sababtoo ah: {{reason}}',
  'Form.submit': 'Dir',
  'Form.title': 'Foomka',
  'Form.unique': 'Qiimahan hadda jiraa ma gaara shuruudda gaarka ahaanshiyaha',
  'Form.validateFailed': 'Xaqiijinta gelitaanka foomka way guuldareystay',
  'Form.nestedError': 'Foomku ma soo muuqan karo saldhig foom ahaan oo kale',
  'Form.rules.message': 'Xaqiijinta wadajirka ee walxaha foomka way guuldareystay',
  'File.imageAfterCrop': 'Sawirka la gooyay',

  // -------------------------------------------------------
  // IFRAME
  // -------------------------------------------------------
  'Iframe.invalid': 'URL-ka iframe waa khaldan yahay',
  'Iframe.invalidProtocol': 'Ma isticmaali kartid URL HTTP ah ee iframe gudaha HTTPS',

  // -------------------------------------------------------
  // SAWIRKA (Image)
  // "Crop" → "Jar" (to cut/trim)
  // -------------------------------------------------------
  'Image.dragTip': 'Jiid si aad u kala soocdo',
  'Image.upload': 'Kor u dhig sawir',
  'Image.errorRetry': 'Kor u dhigistii way guuldareystay, fadlan isku day mar kale',
  'Image.configError': 'Hal keliya oo goynta ama badan ah ayaa la dhigi karaa',
  'Image.crop': 'Jar sawirka',
  'Image.dragDrop': 'Jiid oo dhig sawirrada halkan',
  'Image.height': 'dhererka: {{height}}px',
  'Image.limitMax': 'Cabbirka sawirka ugu yar waa {{info}}',
  'Image.limitMin': 'Cabbirka sawirka ugu badan waa {{info}}',
  'Image.limitRatio': 'Fadlan kor u dhig sawir leh xaddiga {{ration}}',
  'Image.pasteTip': 'Sawir ayaad ka deeqi kartaa xafidaadka',  // clipboard = xafidaad
  'Image.placeholder': 'Guji si aad u doorato sawir ama jiid halkan',
  'Image.size': 'cabbirka: ({{width}}px x {{height}}px)',
  'Image.sizeNotEqual':
    'Sawirka aad doortay kuma buuxinayo shuruudaha cabbirka {{info}}',
  'Image.width': 'ballaarka: {{width}}px',
  'Image.zoomIn': 'Weynee',      // Zoom In = make bigger

  // -------------------------------------------------------
  // DIIWAANKA (Log viewer)
  // -------------------------------------------------------
  'Log.mustHaveSource': 'Isha waa inay ku jirtaa habaynta',
  'Log.showLineNumber': 'Muuji lambarka sadarrada',
  'Log.notShowLineNumber': 'Qari lambarka sadarrada',
  'Log.expand': 'Ballaarinta Hawlgalka',
  'Log.collapse': 'Dhimista Hawlgalka',

  // -------------------------------------------------------
  // GOOBTA (Location)
  // -------------------------------------------------------
  'LocationPicker.placeholder': 'Dooro goobta',
  'LocationPicker.getLocation': 'Guji si aad u hesho macluumaadka goobta',

  // -------------------------------------------------------
  // BISHII (Month)
  // -------------------------------------------------------
  'Month.placeholder': 'Dooro bil',
  'MapPicker.searchPlaceholder': 'Raadi goob',

  // -------------------------------------------------------
  // XIDIDDADA (Navigation)
  // -------------------------------------------------------
  'Nav.sourceError': 'Khalad soo dejista xiriirka',

  // -------------------------------------------------------
  // TIROOYIN (Number Input)
  // -------------------------------------------------------
  'NumberInput.placeholder': 'Fadlan geli tiro',

  // -------------------------------------------------------
  // DOORASHOOYINKA (Options)
  // -------------------------------------------------------
  'Options.addPlaceholder': 'Fadlan geli magac',
  'Options.deleteAPI': 'Waa inuu jiraa API-ga tirtirka',
  'Options.editLabel': 'Wax ka beddel {{label}}',
  'Options.label': 'doorasho',
  'Options.createFailed': 'Abuurida way guuldareystay, fadlan hubi',
  'Options.retry': "Soo dejintu way guuldareystay '{{reason}}', guji si aad mar labaad u tijaabiso",

  // -------------------------------------------------------
  // MEELAHA MADHAN (Placeholders)
  // -------------------------------------------------------
  'placeholder.empty': '<Madhan>',
  'placeholder.enter': 'Geli',
  'placeholder.noData': 'Xog la ma helin',
  'placeholder.noOption': 'Doorasho la ma helin',
  'placeholder.selectData': 'Dooro xogta',

  // -------------------------------------------------------
  // RUBUCII (Quarter)
  // -------------------------------------------------------
  'Quarter.placeholder': 'Dooro rubuc sannadeed',

  // -------------------------------------------------------
  // KU CELCELIN (Repeat)
  // -------------------------------------------------------
  'Repeat.pre': 'Muddo kasta',  // "Per" in context of "Per week/month"

  // -------------------------------------------------------
  // XULASHADA (Select/Checkbox)
  // -------------------------------------------------------
  'Checkboxes.selectAll': 'Dooro/Jooji dhammaan',
  'Select.checkAll': 'Xidh dhammaan',
  'Select.clear': 'Nadiifi',
  'Select.edit': 'Wax ka beddel',
  'Select.upload': 'Dib u kor u dhig',
  'Select.clearAll': 'Nadiifi dhammaan',
  'Select.createLabel': 'Doorasho cusub',
  'Select.placeholder': 'Dooro',
  'Select.searchPromptText': 'Geli si aad u raadiso',
  'Select.selected': 'La doortay',

  // -------------------------------------------------------
  // NIDAAMKA (System messages)
  // -------------------------------------------------------
  'System.error': 'Khalad nidaamka',
  'System.notify': 'Ogeysiiska nidaamka',
  'System.copy': 'Waxa la koobiyeeyay',
  'System.requestError': 'Khalad codsiga: ',
  'System.requestErrorStatus': 'Khalad codsiga, xaaladda koodhka: ',

  // -------------------------------------------------------
  // MIISKA (Table)
  // -------------------------------------------------------
  'Table.addRow': 'Ku dar saf',
  'Table.subAddRow': 'Ku dar saf hoosaad',
  'Table.copyRow': 'Koobiyee saf',
  'Table.columnsVisibility': 'Guji si aad u xukunto muuqaalka tiirarka',
  'Table.columnsSorting': 'Jiid tiirka si aad u kala soocdo',
  'Table.deleteRow': 'Tirtir safka hadda',
  'Table.discard': 'Iska dhaaf',         // Discard = abandon/throw away
  'Table.dragTip': 'Jiid badhanka bidixda si aad u kala soocdo',
  'Table.editing': 'Waa inaad dhamaystirtaa wax ka beddelka',
  'Table.editRow': 'Wax ka beddel safka hadda',
  'Table.modified':
    'Waxaa jira {{modified}} diiwaanno la beddelay, waad awoodaa:',
  'Table.moved': 'Waxaa jira {{moved}} diiwaanno beddelay goobtooda, waad awoodaa:',
  'Table.operation': 'Hawlgal',
  'Table.playload': 'Waa inuu jiraa playload',
  'Table.startSort': 'Guji si aad u bilaabto kala soocida',
  'Table.valueField': 'Waa inuu jiraa valueField',
  'Table.index': 'Tirooyinka',      // Index = numbering/sequence
  'Table.add': 'Ku dar',
  'Table.addButtonDisabledTip':
    'Wax ka beddelka waxa ka dhexeeya, fadlan gudbi marka hore ka hor inta aadan abuurin doorasho cusub',
  'Table.toggleColumn': 'Muuji tiirarka',
  'Table.searchFields': 'Deji goobaha raadinta',

  // -------------------------------------------------------
  // CALAAMADAHA (Tags)
  // "Tag" → "calaamad" (label/mark/tag)
  // -------------------------------------------------------
  'Tag.placeholder': 'Calaamad la ma helin weli',
  'Tag.tip': 'Calaamadaha ugu dambeyntii la isticmaalay',
  'Text.add': 'Cusub {{label}}',

  // -------------------------------------------------------
  // WAKHTIGA (Time)
  // -------------------------------------------------------
  'Time.placeholder': 'Dooro Wakhtiga',
  'TimeNow': 'Hadda',

  // -------------------------------------------------------
  // WAREEJINTA (Transfer - dual-list picker)
  // -------------------------------------------------------
  'Transfer.configError': 'Khalad habaynta',
  'Transfer.refreshIcon': 'Guji si aad u cusboonaysiso',
  'Transfer.searchKeyword': 'Geli erayga raadinta',
  'Transfer.available': 'La heli karo',
  'Transfer.selectd': 'La doortay',
  'Transfer.selectFromLeft': 'Ka dooro bidixda',

  // -------------------------------------------------------
  // GEEDKA (Tree - hierarchical)
  // "node" → "caan" (node/point in a tree structure)
  // -------------------------------------------------------
  'Tree.addChild': 'Ku dar carruur',        // child nodes = carruur (children)
  'Tree.addRoot': 'Ku dar xididka',         // root node = xididka (the root)
  'Tree.editNode': 'Wax ka beddel caan-kan',
  'Tree.invalidArray': 'Data.options, data.items, ama xogtu waa inay ahaadaan arrays',
  'Tree.removeNode': 'Ka saar caan-kan',
  'Tree.root': 'Xididka',

  // -------------------------------------------------------
  // XAQIIJINTA (Validation messages)
  // All "validate.*" keys follow the pattern:
  // "Fadlan geli..." = "Please enter..."
  // "$1" and "$2" are positional placeholders kept as-is
  // -------------------------------------------------------
  'validate.equals': 'Qiimaha waa inuu la mid yahay $1',
  'validate.equalsField': 'Qiimaha waa inuu la mid yahay $1',
  'validate.gt': 'Fadlan geli qiime ka weyn $1',
  'validate.isAlpha': 'Fadlan geli xarfaha',
  'validate.isAlphanumeric': 'Fadlan geli xarfaha ama tirooyin',
  'validate.isEmail': 'Qaabka iimaylku waa khaldan yahay',
  'validate.isFloat': 'Fadlan geli qiime jajaban',          // floating point = qiime jajaban
  'validate.isId': 'Lambarka aqoonsiga ma sax ahayn',
  'validate.isInt': 'Fadlan geli nambar dhammaystiran',     // integer = nambar dhammaystiran (whole number)
  'validate.isJson': 'Qaabka JSON-ku waa khaldan yahay',
  'validate.isLength': 'Fadlan hubi in dhererka waxa ku jira uu yahay $1',
  'validate.isNumeric': 'Fadlan geli tiro',
  'validate.isPhoneNumber': 'Lambarka telefoonka ma sax ahayn',
  'validate.isRequired': 'Tan waa waajib',
  'validate.isTelNumber': 'Lambarka taleefanka ma sax ahayn',
  'validate.isUrl': 'Qaabka URL-ku waa khaldan yahay',
  'validate.isUrlPath': 'Waad kaliya geli kartaa xarfaha, tirooyin, `-` iyo `_` .',
  'validate.isWords': 'Fadlan geli erayo',
  'validate.isZipcode': 'Cinwaanka boostada ma sax ahayn',
  'validate.lt': 'Fadlan geli qiime ka yar $1',
  'validate.matchRegexp':
    'Qaabku ma saxsanayn. Fadlan geli waxa leh xeerka `${1| raw}`.',
  'validate.maximum': 'Qiimaha la geliyay wuxuu ka baxayaa qiimaha ugu badan ee $1',
  'validate.maxLength':
    'Fadlan xukun dhererka waxa ku jira, ha gelin wax ka badan $1 xaraf',
  'validate.minimum': 'Qiimaha la geliyay wuxuu ka hooseeyaa qiimaha ugu yar ee $1',
  'validate.minLength': 'Fadlan wax dheeraad ah geli, ugu yaraan $1 xaraf.',
  'validate.array.minLength': 'Fadlan ku dar xubnaha, ugu yaraan $1 xubno',
  'validate.array.maxLength':
    'Fadlan xukun tirada xubnaha, ma karo in ay ka baxaan $1',
  'validate.isVariableName': 'Fadlan geli magac doorsoomayaa oo sax ah',
  'validate.notEmptyString': 'Fadlan ha gelin xarfaha banaan oo keliya',
  // Taariikhda xaqiijinta - Date validation
  'validate.isDateTimeSame':
    'Qiimaha taariikhda hadda waa khaldan yahay, fadlan geli qiime taariikh ah oo la mid ah $1',
  'validate.isDateTimeBefore':
    'Qiimaha taariikhda hadda waa khaldan yahay, fadlan geli qiime taariikh ah oo ka horreysa $1',
  'validate.isDateTimeAfter':
    'Qiimaha taariikhda hadda waa khaldan yahay, fadlan geli qiime taariikh ah oo ka dambeeysa $1',
  'validate.isDateTimeSameOrBefore':
    'Qiimaha taariikhda hadda waa khaldan yahay, fadlan geli qiime taariikh ah oo la mid ah ama ka horreysa $1',
  'validate.isDateTimeSameOrAfter':
    'Qiimaha taariikhda hadda waa khaldan yahay, fadlan geli qiime taariikh ah oo la mid ah ama ka dambeeysa $1',
  'validate.isDateTimeBetween':
    'Qiimaha taariikhda hadda waa khaldan yahay, fadlan geli qiime taariikh ah oo u dhexeeya $1 iyo $2',
  // Wakhtiga xaqiijinta - Time validation
  'validate.isTimeSame':
    'Qiimaha wakhtiga hadda waa khaldan yahay, fadlan geli qiime wakhti ah oo la mid ah $1',
  'validate.isTimeBefore':
    'Qiimaha wakhtiga hadda waa khaldan yahay, fadlan geli qiime wakhti ah oo ka horreysa $1',
  'validate.isTimeAfter':
    'Qiimaha wakhtiga hadda waa khaldan yahay, fadlan geli qiime wakhti ah oo ka dambeeysa $1',
  'validate.isTimeSameOrBefore':
    'Qiimaha wakhtiga hadda waa khaldan yahay, fadlan geli qiime wakhti ah oo la mid ah ama ka horreysa $1',
  'validate.isTimeSameOrAfter':
    'Qiimaha wakhtiga hadda waa khaldan yahay, fadlan geli qiime wakhti ah oo la mid ah ama ka dambeeysa $1',
  'validate.isTimeBetween':
    'Qiimaha wakhtiga hadda waa khaldan yahay, fadlan geli qiime wakhti ah oo u dhexeeya $1 iyo $2',

  // -------------------------------------------------------
  // HAGAHA (Wizard - multi-step form)
  // -------------------------------------------------------
  'Wizard.configError': 'Khalad habaynta',
  'Wizard.finish': 'Dhamaystir',
  'Wizard.next': 'Xiga',
  'Wizard.prev': 'Hore',
  'Wizard.saveAndNext': 'Keydi & Xiga',

  // -------------------------------------------------------
  // SANADKA XADKA (Year range)
  // -------------------------------------------------------
  'year-to-year': '{{from}} - {{to}}',
  'Year.placeholder': 'Dooro Sannad',

  // -------------------------------------------------------
  // MUUQAALKA (Viewer actions)
  // -------------------------------------------------------
  'rotate': 'Wareejin',
  'rotate.left': 'Bidix u wareejin',
  'rotate.right': 'Midig u wareejin',
  'zoomIn': 'Weynee',
  'zoomOut': 'Yareey',
  'scale.origin': 'Cabbirka asalka ah',
  'Editor.fullscreen': 'Muuqaal buuxa',
  'Editor.exitFullscreen': 'Ka bax muuqaalka buuxa',

  // -------------------------------------------------------
  // XAALADAHA (Condition builder)
  // Logic operators kept close to formal Somali
  // -------------------------------------------------------
  'Condition.not': 'ma aha',
  'Condition.and': 'iyo',
  'Condition.or': 'ama',
  'Condition.collapse': 'ballaar',
  'Condition.add_cond': 'ku dar xaalad',
  'Condition.add_cond_group': 'ku dar koox xaalad',
  'Condition.delete_cond_group': 'tirtir koox xaalad',
  'Condition.equal': 'la mid ah',
  'Condition.not_equal': 'la mid ma aha',
  'Condition.less': 'ka yar',
  'Condition.less_or_equal': 'ka yar ama la mid ah',
  'Condition.greater': 'ka weyn',
  'Condition.greater_or_equal': 'ka weyn ama la mid ah',
  'Condition.between': 'u dhexeeya',
  'Condition.not_between': 'u dhexayn',
  'Condition.is_empty': 'madhan',
  'Condition.is_not_empty': 'ma madan',
  'Condition.like': 'ku jira',
  'Condition.not_like': 'ku jirin',
  'Condition.starts_with': 'ka bilaabma',
  'Condition.ends_with': 'ku dhammaanaysa',
  'Condition.select_equals': 'la mid ah',
  'Condition.select_not_equals': 'la mid ma aha',
  'Condition.select_any_in': 'ku jira',
  'Condition.select_not_any_in': 'ku jirin',
  'Condition.placeholder': 'Fadlan geli qoraal',
  'Condition.cond_placeholder': 'dooro xaalad',
  'Condition.field_placeholder': 'dooro goob',
  'Condition.blank': 'madhan',
  'Condition.expression': 'muujin',        // expression = mathematical/logical expression
  'Condition.formula_placeholder': 'Fadlan geli qaacidada',
  'Condition.fun_error': 'Hawlwadeenku ma qeexi ahayn',
  'Condition.configured': 'La habeeyay',
  'Condition.isRequired': 'Xaaladdu waa waajib',

  // -------------------------------------------------------
  // MIISKA GELINTA (InputTable)
  // -------------------------------------------------------
  'InputTable.uniqueError': 'Tiirka `{{label}}` xaqiijinta gaarka ahaanshiyaha way guuldareystay',

  // -------------------------------------------------------
  // WAKHTILAHA (Timeline)
  // -------------------------------------------------------
  'Timeline.collapseText': 'Ballaar',
  'Timeline.expandText': 'Laab',

  // -------------------------------------------------------
  // BALLAARINTA / LAABASHADA (Expand/Collapse general)
  // -------------------------------------------------------
  'collapse': 'Laab',
  'expand': 'Ballaar',

  // -------------------------------------------------------
  // QAACIDADA TIFATIRAHA (Formula Editor)
  // "Formula" → "Qaacidada" (mathematical formula/rule)
  // -------------------------------------------------------
  'FormulaEditor.btnLabel': 'Tifatire Qaacidada',
  'FormulaEditor.title': 'Tifatiraha Qaacidada',
  'FormulaEditor.run': 'Orod',
  'FormulaEditor.sourceMode': 'Habka Isha',
  'FormulaEditor.runContext': 'Xaaladda Orodka',
  'FormulaEditor.runResult': 'Natiijada Orodka',
  'FormulaEditor.toggleAll': 'Ballaar Dhammaan',
  'FormulaEditor.variable': 'Doorsoomayaa',      // variable = doorsoomayaa (changeable thing)
  'FormulaEditor.function': 'Hawlwadeenka',      // function = hawlwadeenka (function/operator)
  'FormulaEditor.invalidData': 'Xogta waa khaldan tahay, goobta ama sababta waa {{err}}',
  'FormulaEditor.invalidValue': 'Qiimaha waa khaldan yahay, sababtu waa qaab khalad ah',

  // -------------------------------------------------------
  // SOO JIIDASHADA (Pull to Refresh)
  // -------------------------------------------------------
  'pullRefresh.normalText': '',
  'pullRefresh.pullingText': 'Hoos u jiid si aad u cusboonaysiso...',
  'pullRefresh.loosingText': 'Sii si aad u cusboonaysiso...',
  'pullRefresh.loadingText': 'Waa la rarayo...',
  'pullRefresh.successText': 'Rarista wey guulaystay',
  'pullRefresh.completedText': 'Xog kale la ma helin',
  'pullRefresh.crud2NormalText': 'Kor u jiid si aad wax badan u rariso',
  'pullRefresh.crud2PullingText': 'Kor u jiid si aad wax badan u rariso',
  'pullRefresh.crud2LoosingText': 'Sii si aad u rariso',

  // -------------------------------------------------------
  // DOORASHADA (Picker)
  // -------------------------------------------------------
  'Picker.placeholder': 'Guji calaamadda dhanka midig',

  // -------------------------------------------------------
  // XULASHADA ISTICMAALAHA (User Select)
  // -------------------------------------------------------
  'UserSelect.edit': 'wax ka beddel',
  'UserSelect.save': 'keydi',
  'UserSelect.resultSort': 'Kala sooc natiijada xulashada',
  'UserSelect.selected': 'La doortay',
  'UserSelect.clear': 'maran',
  'UserSelect.sure': 'dir',

  // -------------------------------------------------------
  // NOOCYADA XOGTA (Schema Types)
  // Technical data types - kept recognizable where possible
  // -------------------------------------------------------
  'SchemaType.string': 'Qoraal',
  'SchemaType.number': 'Tiro',
  'SchemaType.integer': 'Nambar dhammaystiran',
  'SchemaType.object': 'Walax',       // object = walax (thing/object)
  'SchemaType.array': 'Liis',         // array = liis (list)
  'SchemaType.boolean': 'Haa/Maya',   // boolean = yes/no
  'SchemaType.any': 'Kasta',          // any = any type
  'SchemaType.null': 'Maran',         // null = empty/nothing

  // -------------------------------------------------------
  // QAABKA JSON (JSON Schema editor)
  // -------------------------------------------------------
  'JSONSchema.type': 'Nooca',
  'JSONSchema.required': 'Waajib',
  'JSONSchema.title': 'Cinwaanka',
  'JSONSchema.default': 'Caadiga',
  'JSONSchema.description': 'Sharaxaadda',
  'JSONSchema.key': 'Furaha',         // key = furaha (the key/opener)
  'JSONSchema.array_items': 'Walxaha',
  'JSONSchema.members': 'Xubnahaa',
  'JSONSchema.key_duplicated': 'Furaha hore ayuu jiraa',
  'JSONSchema.key_invalid': 'Furaha waa khaldan yahay',

  // -------------------------------------------------------
  // CALAAMADA ASTAANTA (Icon Select)
  // -------------------------------------------------------
  'IconSelect.all': 'Dhammaan',
  'IconSelect.choice': 'Xulashada calaamadaha',

  // -------------------------------------------------------
  // TALLAABOOYINKA (Steps)
  // -------------------------------------------------------
  'Steps.step': 'Tallaabo {{index}}',

  // -------------------------------------------------------
  // QIIMEYAASHA QAACIDADA (Formula Input)
  // -------------------------------------------------------
  'FormulaInput.True': 'Haa',    // True = Haa (Yes/True in Somali)
  'FormulaInput.False': 'Maya',  // False = Maya (No/False in Somali)

  // -------------------------------------------------------
  // SAXIIXA (Signature pad)
  // -------------------------------------------------------
  'Signature.clear': 'nadiifi',
  'Signature.undo': 'dib u noqo',
  'Signature.confirm': 'xaqiiji',
  'Signature.cancel': 'jooji',
  'Signature.embedLabel': 'Guji si aad u saxiixdo',
  'Signature.embedUpdateLabel': 'Cusboonaysii si aad u saxiixdo',

  // -------------------------------------------------------
  // KOODHKA QR (QR Code)
  // -------------------------------------------------------
  'QRCode.tooLong':
    'Qiimaha koodhka QR-ku aad buu u dheer yahay, fadlan deji qoraalka inuu ka hooseeyaa {{max}} xaraf.',

  // -------------------------------------------------------
  // BEDDELKA (Switch toggle)
  // -------------------------------------------------------
  'swith.on': 'Fur',    // On = open/turn on
  'swith.off': 'Xidh'  // Off = close/turn off
});
