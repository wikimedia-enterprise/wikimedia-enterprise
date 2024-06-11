package wmf

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

const errorResponse = `{
	"error": {
		"code": "badvalue",
		"info": "Unrecognized value for parameter \"apfilterredir\": asdas.",
		"docref": "See https://en.wikipedia.org/w/api.php for API usage. Subscribe to the mediawiki-api-announce mailing list at &lt;https://lists.wikimedia.org/postorius/lists/mediawiki-api-announce.lists.wikimedia.org/&gt; for notice of API deprecations and breaking changes."
	},
	"servedby": "mw1426"
}`

const getPagesResponse = `{
	"batchcomplete":true,
	"query":{
		"pages":[
			{"pageid":9228,"ns":0,"title":"Earth","contentmodel":"wikitext","pagelanguage":"en","pagelanguagehtmlcode":"en","pagelanguagedir":"ltr","touched":"2022-11-07T09:53:09Z","lastrevid":1119965032,"length":190421,"protection":[{"type":"edit","level":"autoconfirmed","expiry":"infinity"},{"type":"move","level":"sysop","expiry":"infinity"}],"restrictiontypes":["edit","move"],"fullurl":"https://en.wikipedia.org/wiki/Earth","editurl":"https://en.wikipedia.org/w/index.php?title=Earth&action=edit","canonicalurl":"https://en.wikipedia.org/wiki/Earth","displaytitle":"Earth","revisions":[{"revid":1119965032,"parentid":1119594067,"minor":true,"user":"Finnusertop","userid":19089174,"timestamp":"2022-11-04T10:47:24Z","slots":{"main":{"contentmodel":"wikitext","contentformat":"text/x-wiki","content":"...wikitext..."}},"comment":"/*References*/cs1","tags":["wikieditor"],"oresscores":{"damaging":{"true":0.011,"false":0.989},"goodfaith":{"true":0.995,"false":0.0050000000000000044},"articlequality":{"Stub":0.752}}}],"wbentityusage":{"Q2":{"aspects":["C","D.en","O","S","T"]}},"pageprops":{"wikibase_item":"Q2"},"redirects":[{"pageid":9215,"ns":0,"title":"EartH"},{"pageid":307601,"ns":0,"title":"Sol3"},{"pageid":603544,"ns":0,"title":"TheEarth"},{"pageid":896072,"ns":0,"title":"Earth(Planet)"},{"pageid":1096589,"ns":0,"title":"Surfaceareaofearth"},{"pageid":1191327,"ns":0,"title":"Terra(planet)"},{"pageid":1324754,"ns":0,"title":"Theplanetearth"},{"pageid":1415438,"ns":0,"title":"Terra(namefortheearth)"},{"pageid":1788541,"ns":0,"title":"LocalPlanet"},{"pageid":2237401,"ns":0,"title":"ThirdPlanet"},{"pageid":2742548,"ns":0,"title":"Globe(Earth)"},{"pageid":3520701,"ns":0,"title":"ThirdplanetfromtheSun"},{"pageid":3601947,"ns":0,"title":"Tellus(Planet)"},{"pageid":4476832,"ns":0,"title":"SolIII"},{"pageid":5222588,"ns":0,"title":"Planetearth"},{"pageid":5423363,"ns":0,"title":"World(geography)"},{"pageid":8678510,"ns":0,"title":"Earth(planet)"},{"pageid":9090641,"ns":0,"title":"ThePlanetEarth"},{"pageid":9455987,"ns":0,"title":"HomePlanet"},{"pageid":9898684,"ns":0,"title":"Earth(word)"},{"pageid":13181153,"ns":0,"title":"MeandensityoftheEarth"},{"pageid":13935837,"ns":0,"title":"Eareth"},{"pageid":15203095,"ns":0,"title":"Blueandgreenplanet"},{"pageid":15203100,"ns":0,"title":"Greenandblueplanet"},{"pageid":16430764,"ns":0,"title":"Earth’ssurface"},{"pageid":16972296,"ns":0,"title":"Earth'ssurface"},{"pageid":18755374,"ns":0,"title":"PlanetofWater"},{"pageid":19790623,"ns":0,"title":"Sol-3"},{"pageid":20384608,"ns":0,"title":"EARTH"},{"pageid":22759962,"ns":0,"title":"Thirdplanet"},{"pageid":23775266,"ns":0,"title":"Earth'smeandensity"},{"pageid":26366190,"ns":0,"title":"CompositionoftheEarth"},{"pageid":27384837,"ns":0,"title":"Telluris"},{"pageid":27706257,"ns":0,"title":"SolPrime"},{"pageid":28257717,"ns":0,"title":"LexicographyofEarth"},{"pageid":31193038,"ns":0,"title":"Earth,Sol"},{"pageid":33364470,"ns":0,"title":"FormationoftheEarth"},{"pageid":33810062,"ns":0,"title":"Etymologyoftheword\"Earth\""},{"pageid":35531228,"ns":0,"title":"SurfaceoftheEarth"},{"pageid":43507855,"ns":0,"title":"Tierra(planet)"},{"pageid":43822591,"ns":0,"title":"3rdplanet"},{"pageid":47103485,"ns":0,"title":"806.4616.0110"},{"pageid":48120239,"ns":0,"title":"PlanetTerra"},{"pageid":56078981,"ns":0,"title":"TheplanetEarth"},{"pageid":56079851,"ns":0,"title":"PlanetEarth"},{"pageid":57857323,"ns":0,"title":"Sizeoftheearth"},{"pageid":58783959,"ns":0,"title":"Theearth"},{"pageid":63738768,"ns":0,"title":"Earthsurface"},{"pageid":64715694,"ns":0,"title":"Earth'sdensity"},{"pageid":64715695,"ns":0,"title":"DensityoftheEarth"},{"pageid":66075911,"ns":0,"title":"PlanetThree"},{"pageid":67560020,"ns":0,"title":"ClimateofEarth"},{"pageid":67713162,"ns":118,"title":"Draft:Earth"},{"pageid":68559411,"ns":0,"title":"FormationofEarth"}]},
			{"pageid":46396,"ns":0,"title":"Ninja","contentmodel":"wikitext","pagelanguage":"en","pagelanguagehtmlcode":"en","pagelanguagedir":"ltr","touched":"2022-11-07T21:42:13Z","lastrevid":1117662369,"length":79921,"protection":[{"type":"edit","level":"autoconfirmed","expiry":"infinity"},{"type":"move","level":"autoconfirmed","expiry":"infinity"}],"restrictiontypes":["edit","move"],"fullurl":"https://en.wikipedia.org/wiki/Ninja","editurl":"https://en.wikipedia.org/w/index.php?title=Ninja&action=edit","canonicalurl":"https://en.wikipedia.org/wiki/Ninja","displaytitle":"Ninja","revisions":[{"revid":1117662369,"parentid":1117662080,"minor":false,"user":"ToastforTeddy","userid":44098681,"timestamp":"2022-10-22T23:00:58Z","slots":{"main":{"contentmodel":"wikitext","contentformat":"text/x-wiki","content":"...wikitext..."}},"comment":"Contentsofsectionmovedto==Seealso==","tags":["visualeditor-wikitext"],"oresscores":{"damaging":{"true":0.268,"false":0.732},"goodfaith":{"true":0.923,"false":0.07699999999999996},"articlequality":{"Stub":0.747}}}],"wbentityusage":{"Q7430520":{"aspects":["S"]},"Q9402":{"aspects":["C","D.en","O","S","T"]}},"pageprops":{"wikibase_item":"Q9402"},"redirects":[{"pageid":340047,"ns":0,"title":"Ninjas"},{"pageid":597512,"ns":0,"title":"Ninzya"},{"pageid":617095,"ns":0,"title":"Sinobi"},{"pageid":1799908,"ns":0,"title":"Shinobishōzoku"},{"pageid":3088559,"ns":0,"title":"Nukenin"},{"pageid":3562652,"ns":0,"title":"Historyoftheninja"},{"pageid":3563229,"ns":0,"title":"HistoryoftheNinja"},{"pageid":5032477,"ns":0,"title":"Shinobishozoku"},{"pageid":5110667,"ns":0,"title":"Jonin"},{"pageid":5355512,"ns":0,"title":"ShinobiShozoku"},{"pageid":6796660,"ns":0,"title":"Chunin"},{"pageid":10690516,"ns":0,"title":"Shinobi"},{"pageid":15203846,"ns":0,"title":"忍者"},{"pageid":17425138,"ns":0,"title":"Suppa"},{"pageid":22173937,"ns":0,"title":"NINJA"},{"pageid":22248985,"ns":0,"title":"Shinobishozoku"},{"pageid":24163001,"ns":0,"title":"Shinobi-no-mono"},{"pageid":56504885,"ns":0,"title":"Chūnin"},{"pageid":59779622,"ns":0,"title":"忍び"},{"pageid":62995468,"ns":0,"title":"🥷"},{"pageid":63676735,"ns":0,"title":"🥷🏻"},{"pageid":63676737,"ns":0,"title":"🥷🏼"},{"pageid":63676738,"ns":0,"title":"🥷🏽"},{"pageid":63676740,"ns":0,"title":"🥷🏾"},{"pageid":63676741,"ns":0,"title":"🥷🏿"},{"pageid":70535257,"ns":0,"title":"Goshiki-mai"},{"pageid":70535258,"ns":0,"title":"Goshiki-Mai"}]}
		]
	}
}`

const (
	getLanguagesResponse = `{"sitematrix":{"count":996,"0":{"code":"aa","name":"Qafár af","site":[{"url":"https://aa.wikipedia.org","dbname":"aawiki","code":"wiki","sitename":"Wikipedia","closed":true},{"url":"https://aa.wiktionary.org","dbname":"aawiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true},{"url":"https://aa.wikibooks.org","dbname":"aawikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Afar"},"1":{"code":"ab","name":"аԥсшәа","site":[{"url":"https://ab.wikipedia.org","dbname":"abwiki","code":"wiki","sitename":"Авикипедиа"},{"url":"https://ab.wiktionary.org","dbname":"abwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true}],"dir":"ltr","localname":"Abkasies"},"2":{"code":"ace","name":"Acèh","site":[{"url":"https://ace.wikipedia.org","dbname":"acewiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Atsjenees"},"3":{"code":"ady","name":"адыгабзэ","site":[{"url":"https://ady.wikipedia.org","dbname":"adywiki","code":"wiki","sitename":"Википедие"}],"dir":"ltr","localname":"Adyghe"},"4":{"code":"af","name":"Afrikaans","site":[{"url":"https://af.wikipedia.org","dbname":"afwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://af.wiktionary.org","dbname":"afwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://af.wikibooks.org","dbname":"afwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://af.wikiquote.org","dbname":"afwikiquote","code":"wikiquote","sitename":"Wikiquote"}],"dir":"ltr","localname":"Afrikaans"},"5":{"code":"ak","name":"Akan","site":[{"url":"https://ak.wikipedia.org","dbname":"akwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ak.wiktionary.org","dbname":"akwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true},{"url":"https://ak.wikibooks.org","dbname":"akwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Akan"},"6":{"code":"als","name":"Alemannisch","site":[{"url":"https://als.wikipedia.org","dbname":"alswiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Alemannisch"},"7":{"code":"alt","name":"алтай тил","site":[{"url":"https://alt.wikipedia.org","dbname":"altwiki","code":"wiki","sitename":"Википедия"}],"dir":"ltr","localname":"Suid-Altai"},"8":{"code":"am","name":"አማርኛ","site":[{"url":"https://am.wikipedia.org","dbname":"amwiki","code":"wiki","sitename":"ውክፔዲያ"},{"url":"https://am.wiktionary.org","dbname":"amwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://am.wikiquote.org","dbname":"amwikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Amharies"},"9":{"code":"ami","name":"Pangcah","site":[{"url":"https://ami.wikipedia.org","dbname":"amiwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Amis"},"10":{"code":"an","name":"aragonés","site":[{"url":"https://an.wikipedia.org","dbname":"anwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://an.wiktionary.org","dbname":"anwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Aragonees"},"11":{"code":"ang","name":"Ænglisc","site":[{"url":"https://ang.wikipedia.org","dbname":"angwiki","code":"wiki","sitename":"Wikipǣdia"},{"url":"https://ang.wiktionary.org","dbname":"angwiktionary","code":"wiktionary","sitename":"Wikiwordbōc"},{"url":"https://ang.wikibooks.org","dbname":"angwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://ang.wikiquote.org","dbname":"angwikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true},{"url":"https://ang.wikisource.org","dbname":"angwikisource","code":"wikisource","sitename":"Wicifruma","closed":true}],"dir":"ltr","localname":"Old English"},"12":{"code":"ar","name":"العربية","site":[{"url":"https://ar.wikipedia.org","dbname":"arwiki","code":"wiki","sitename":"ويكيبيديا"},{"url":"https://ar.wiktionary.org","dbname":"arwiktionary","code":"wiktionary","sitename":"ويكاموس"},{"url":"https://ar.wikibooks.org","dbname":"arwikibooks","code":"wikibooks","sitename":"ويكي_الكتب"},{"url":"https://ar.wikinews.org","dbname":"arwikinews","code":"wikinews","sitename":"ويكي_الأخبار"},{"url":"https://ar.wikiquote.org","dbname":"arwikiquote","code":"wikiquote","sitename":"ويكي_الاقتباس"},{"url":"https://ar.wikisource.org","dbname":"arwikisource","code":"wikisource","sitename":"ويكي_مصدر"},{"url":"https://ar.wikiversity.org","dbname":"arwikiversity","code":"wikiversity","sitename":"ويكي الجامعة"}],"dir":"rtl","localname":"Arabies"},"13":{"code":"arc","name":"ܐܪܡܝܐ","site":[{"url":"https://arc.wikipedia.org","dbname":"arcwiki","code":"wiki","sitename":"ܘܝܩܝܦܕܝܐ"}],"dir":"rtl","localname":"Aramees"},"14":{"code":"ary","name":"الدارجة","site":[{"url":"https://ary.wikipedia.org","dbname":"arywiki","code":"wiki","sitename":"ويكيپيديا"}],"dir":"rtl","localname":"Moroccan Arabic"},"15":{"code":"arz","name":"مصرى","site":[{"url":"https://arz.wikipedia.org","dbname":"arzwiki","code":"wiki","sitename":"ويكيبيديا"}],"dir":"rtl","localname":"Egyptian Arabic"},"16":{"code":"as","name":"অসমীয়া","site":[{"url":"https://as.wikipedia.org","dbname":"aswiki","code":"wiki","sitename":"অসমীয়া ৱিকিপিডিয়া"},{"url":"https://as.wiktionary.org","dbname":"aswiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true},{"url":"https://as.wikibooks.org","dbname":"aswikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://as.wikiquote.org","dbname":"aswikiquote","code":"wikiquote","sitename":"ৱিকিউদ্ধৃতি"},{"url":"https://as.wikisource.org","dbname":"aswikisource","code":"wikisource","sitename":"ৱিকিউৎস"}],"dir":"ltr","localname":"Assamees"},"17":{"code":"ast","name":"asturianu","site":[{"url":"https://ast.wikipedia.org","dbname":"astwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ast.wiktionary.org","dbname":"astwiktionary","code":"wiktionary","sitename":"Wikcionariu"},{"url":"https://ast.wikibooks.org","dbname":"astwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://ast.wikiquote.org","dbname":"astwikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Asturies"},"18":{"code":"atj","name":"Atikamekw","site":[{"url":"https://atj.wikipedia.org","dbname":"atjwiki","code":"wiki","sitename":"Wikipetcia"}],"dir":"ltr","localname":"Atikamekw"},"19":{"code":"av","name":"авар","site":[{"url":"https://av.wikipedia.org","dbname":"avwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://av.wiktionary.org","dbname":"avwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true}],"dir":"ltr","localname":"Avaries"},"20":{"code":"avk","name":"Kotava","site":[{"url":"https://avk.wikipedia.org","dbname":"avkwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Kotava"},"21":{"code":"awa","name":"अवधी","site":[{"url":"https://awa.wikipedia.org","dbname":"awawiki","code":"wiki","sitename":"विकिपीडिया"}],"dir":"ltr","localname":"Awadhi"},"22":{"code":"ay","name":"Aymar aru","site":[{"url":"https://ay.wikipedia.org","dbname":"aywiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ay.wiktionary.org","dbname":"aywiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://ay.wikibooks.org","dbname":"aywikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Aymara"},"23":{"code":"az","name":"azərbaycanca","site":[{"url":"https://az.wikipedia.org","dbname":"azwiki","code":"wiki","sitename":"Vikipediya"},{"url":"https://az.wiktionary.org","dbname":"azwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://az.wikibooks.org","dbname":"azwikibooks","code":"wikibooks","sitename":"Vikikitab"},{"url":"https://az.wikiquote.org","dbname":"azwikiquote","code":"wikiquote","sitename":"Vikisitat"},{"url":"https://az.wikisource.org","dbname":"azwikisource","code":"wikisource","sitename":"Vikimənbə"}],"dir":"ltr","localname":"Azerbeidjans"},"24":{"code":"azb","name":"تۆرکجه","site":[{"url":"https://azb.wikipedia.org","dbname":"azbwiki","code":"wiki","sitename":"ویکی‌پدیا"}],"dir":"rtl","localname":"South Azerbaijani"},"25":{"code":"ba","name":"башҡортса","site":[{"url":"https://ba.wikipedia.org","dbname":"bawiki","code":"wiki","sitename":"Википедия"},{"url":"https://ba.wikibooks.org","dbname":"bawikibooks","code":"wikibooks","sitename":"Викидәреслек"}],"dir":"ltr","localname":"Baskir"},"26":{"code":"ban","name":"Basa Bali","site":[{"url":"https://ban.wikipedia.org","dbname":"banwiki","code":"wiki","sitename":"Wikipédia"},{"url":"https://ban.wikisource.org","dbname":"banwikisource","code":"wikisource","sitename":"Wikisource"}],"dir":"ltr","localname":"Balinees"},"27":{"code":"bar","name":"Boarisch","site":[{"url":"https://bar.wikipedia.org","dbname":"barwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Bavarian"},"28":{"code":"bat-smg","name":"žemaitėška","site":[{"url":"https://bat-smg.wikipedia.org","dbname":"bat_smgwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Samogitian"},"29":{"code":"bcl","name":"Bikol Central","site":[{"url":"https://bcl.wikipedia.org","dbname":"bclwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://bcl.wiktionary.org","dbname":"bclwiktionary","code":"wiktionary","sitename":"Wiksyunaryo"},{"url":"https://bcl.wikiquote.org","dbname":"bclwikiquote","code":"wikiquote","sitename":"Wikiquote"}],"dir":"ltr","localname":"Central Bikol"},"30":{"code":"be","name":"беларуская","site":[{"url":"https://be.wikipedia.org","dbname":"bewiki","code":"wiki","sitename":"Вікіпедыя"},{"url":"https://be.wiktionary.org","dbname":"bewiktionary","code":"wiktionary","sitename":"Вікіслоўнік"},{"url":"https://be.wikibooks.org","dbname":"bewikibooks","code":"wikibooks","sitename":"Вікікнігі"},{"url":"https://be.wikiquote.org","dbname":"bewikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://be.wikisource.org","dbname":"bewikisource","code":"wikisource","sitename":"Вікікрыніцы"}],"dir":"ltr","localname":"Belarussies"},"31":{"code":"be-tarask","name":"беларуская (тарашкевіца)","site":[],"dir":"ltr","localname":"Belarusian (Taraškievica orthography)"},"32":{"code":"be-x-old","name":"беларуская (тарашкевіца)","site":[{"url":"https://be-tarask.wikipedia.org","dbname":"be_x_oldwiki","code":"wiki","sitename":"Вікіпэдыя"}],"dir":"ltr","localname":"Belarusian (Taraškievica orthography)"},"33":{"code":"bg","name":"български","site":[{"url":"https://bg.wikipedia.org","dbname":"bgwiki","code":"wiki","sitename":"Уикипедия"},{"url":"https://bg.wiktionary.org","dbname":"bgwiktionary","code":"wiktionary","sitename":"Уикиречник"},{"url":"https://bg.wikibooks.org","dbname":"bgwikibooks","code":"wikibooks","sitename":"Уикикниги"},{"url":"https://bg.wikinews.org","dbname":"bgwikinews","code":"wikinews","sitename":"Уикиновини","closed":true},{"url":"https://bg.wikiquote.org","dbname":"bgwikiquote","code":"wikiquote","sitename":"Уикицитат"},{"url":"https://bg.wikisource.org","dbname":"bgwikisource","code":"wikisource","sitename":"Уикиизточник"}],"dir":"ltr","localname":"Bulgaars"},"34":{"code":"bh","name":"भोजपुरी","site":[{"url":"https://bh.wikipedia.org","dbname":"bhwiki","code":"wiki","sitename":"विकिपीडिया"},{"url":"https://bh.wiktionary.org","dbname":"bhwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true}],"dir":"ltr","localname":"Bhojpuri"},"35":{"code":"bi","name":"Bislama","site":[{"url":"https://bi.wikipedia.org","dbname":"biwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://bi.wiktionary.org","dbname":"biwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true},{"url":"https://bi.wikibooks.org","dbname":"biwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Bislama"},"36":{"code":"bjn","name":"Banjar","site":[{"url":"https://bjn.wikipedia.org","dbname":"bjnwiki","code":"wiki","sitename":"Wikipidia"},{"url":"https://bjn.wiktionary.org","dbname":"bjnwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Banjar"},"37":{"code":"blk","name":"ပအိုဝ်ႏဘာႏသာႏ","site":[{"url":"https://blk.wikipedia.org","dbname":"blkwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Pa'O"},"38":{"code":"bm","name":"bamanankan","site":[{"url":"https://bm.wikipedia.org","dbname":"bmwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://bm.wiktionary.org","dbname":"bmwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true},{"url":"https://bm.wikibooks.org","dbname":"bmwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://bm.wikiquote.org","dbname":"bmwikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Bambara"},"39":{"code":"bn","name":"বাংলা","site":[{"url":"https://bn.wikipedia.org","dbname":"bnwiki","code":"wiki","sitename":"উইকিপিডিয়া"},{"url":"https://bn.wiktionary.org","dbname":"bnwiktionary","code":"wiktionary","sitename":"উইকিঅভিধান"},{"url":"https://bn.wikibooks.org","dbname":"bnwikibooks","code":"wikibooks","sitename":"উইকিবই"},{"url":"https://bn.wikiquote.org","dbname":"bnwikiquote","code":"wikiquote","sitename":"উইকিউক্তি"},{"url":"https://bn.wikisource.org","dbname":"bnwikisource","code":"wikisource","sitename":"উইকিসংকলন"},{"url":"https://bn.wikivoyage.org","dbname":"bnwikivoyage","code":"wikivoyage","sitename":"উইকিভ্রমণ"}],"dir":"ltr","localname":"Bengaals"},"40":{"code":"bo","name":"བོད་ཡིག","site":[{"url":"https://bo.wikipedia.org","dbname":"bowiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://bo.wiktionary.org","dbname":"bowiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true},{"url":"https://bo.wikibooks.org","dbname":"bowikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Tibettaans"},"41":{"code":"bpy","name":"বিষ্ণুপ্রিয়া মণিপুরী","site":[{"url":"https://bpy.wikipedia.org","dbname":"bpywiki","code":"wiki","sitename":"উইকিপিডিয়া"}],"dir":"ltr","localname":"Bishnupriya"},"42":{"code":"br","name":"brezhoneg","site":[{"url":"https://br.wikipedia.org","dbname":"brwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://br.wiktionary.org","dbname":"brwiktionary","code":"wiktionary","sitename":"Wikeriadur"},{"url":"https://br.wikiquote.org","dbname":"brwikiquote","code":"wikiquote","sitename":"Wikiarroud"},{"url":"https://br.wikisource.org","dbname":"brwikisource","code":"wikisource","sitename":"Wikimammenn"}],"dir":"ltr","localname":"Bretons"},"43":{"code":"bs","name":"bosanski","site":[{"url":"https://bs.wikipedia.org","dbname":"bswiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://bs.wiktionary.org","dbname":"bswiktionary","code":"wiktionary","sitename":"Wikirječnik"},{"url":"https://bs.wikibooks.org","dbname":"bswikibooks","code":"wikibooks","sitename":"Wikiknjige"},{"url":"https://bs.wikinews.org","dbname":"bswikinews","code":"wikinews","sitename":"Wikivijesti"},{"url":"https://bs.wikiquote.org","dbname":"bswikiquote","code":"wikiquote","sitename":"Wikicitati"},{"url":"https://bs.wikisource.org","dbname":"bswikisource","code":"wikisource","sitename":"Wikizvor"}],"dir":"ltr","localname":"Bosnies"},"44":{"code":"bug","name":"ᨅᨔ ᨕᨘᨁᨗ","site":[{"url":"https://bug.wikipedia.org","dbname":"bugwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Buginees"},"45":{"code":"bxr","name":"буряад","site":[{"url":"https://bxr.wikipedia.org","dbname":"bxrwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Russia Buriat"},"46":{"code":"ca","name":"català","site":[{"url":"https://ca.wikipedia.org","dbname":"cawiki","code":"wiki","sitename":"Viquipèdia"},{"url":"https://ca.wiktionary.org","dbname":"cawiktionary","code":"wiktionary","sitename":"Viccionari"},{"url":"https://ca.wikibooks.org","dbname":"cawikibooks","code":"wikibooks","sitename":"Viquillibres"},{"url":"https://ca.wikinews.org","dbname":"cawikinews","code":"wikinews","sitename":"Viquinotícies"},{"url":"https://ca.wikiquote.org","dbname":"cawikiquote","code":"wikiquote","sitename":"Viquidites"},{"url":"https://ca.wikisource.org","dbname":"cawikisource","code":"wikisource","sitename":"Viquitexts"}],"dir":"ltr","localname":"Katalaans"},"47":{"code":"cbk-zam","name":"Chavacano de Zamboanga","site":[{"url":"https://cbk-zam.wikipedia.org","dbname":"cbk_zamwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Chavacano"},"48":{"code":"cdo","name":"閩東語 / Mìng-dĕ̤ng-ngṳ̄","site":[{"url":"https://cdo.wikipedia.org","dbname":"cdowiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Min Dong Chinese"},"49":{"code":"ce","name":"нохчийн","site":[{"url":"https://ce.wikipedia.org","dbname":"cewiki","code":"wiki","sitename":"Википеди"}],"dir":"ltr","localname":"Tsjetsjeens"},"50":{"code":"ceb","name":"Cebuano","site":[{"url":"https://ceb.wikipedia.org","dbname":"cebwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Cebuano"},"51":{"code":"ch","name":"Chamoru","site":[{"url":"https://ch.wikipedia.org","dbname":"chwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ch.wiktionary.org","dbname":"chwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true},{"url":"https://ch.wikibooks.org","dbname":"chwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Chamorro"},"52":{"code":"cho","name":"Chahta Anumpa","site":[{"url":"https://cho.wikipedia.org","dbname":"chowiki","code":"wiki","sitename":"Wikipedia","closed":true}],"dir":"ltr","localname":"Choctaw"},"53":{"code":"chr","name":"ᏣᎳᎩ","site":[{"url":"https://chr.wikipedia.org","dbname":"chrwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://chr.wiktionary.org","dbname":"chrwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Cherokees"},"54":{"code":"chy","name":"Tsetsêhestâhese","site":[{"url":"https://chy.wikipedia.org","dbname":"chywiki","code":"wiki","sitename":"Tsétsêhéstâhese Wikipedia"}],"dir":"ltr","localname":"Cheyennees"},"55":{"code":"ckb","name":"کوردی","site":[{"url":"https://ckb.wikipedia.org","dbname":"ckbwiki","code":"wiki","sitename":"ویکیپیدیا"}],"dir":"rtl","localname":"Sorani"},"56":{"code":"co","name":"corsu","site":[{"url":"https://co.wikipedia.org","dbname":"cowiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://co.wiktionary.org","dbname":"cowiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://co.wikibooks.org","dbname":"cowikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://co.wikiquote.org","dbname":"cowikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Korsikaans"},"57":{"code":"cr","name":"Nēhiyawēwin / ᓀᐦᐃᔭᐍᐏᐣ","site":[{"url":"https://cr.wikipedia.org","dbname":"crwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://cr.wiktionary.org","dbname":"crwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true},{"url":"https://cr.wikiquote.org","dbname":"crwikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Cree"},"58":{"code":"crh","name":"qırımtatarca","site":[{"url":"https://crh.wikipedia.org","dbname":"crhwiki","code":"wiki","sitename":"Vikipediya"}],"dir":"ltr","localname":"Crimean Tatar"},"59":{"code":"cs","name":"čeština","site":[{"url":"https://cs.wikipedia.org","dbname":"cswiki","code":"wiki","sitename":"Wikipedie"},{"url":"https://cs.wiktionary.org","dbname":"cswiktionary","code":"wiktionary","sitename":"Wikislovník"},{"url":"https://cs.wikibooks.org","dbname":"cswikibooks","code":"wikibooks","sitename":"Wikiknihy"},{"url":"https://cs.wikinews.org","dbname":"cswikinews","code":"wikinews","sitename":"Wikizprávy"},{"url":"https://cs.wikiquote.org","dbname":"cswikiquote","code":"wikiquote","sitename":"Wikicitáty"},{"url":"https://cs.wikisource.org","dbname":"cswikisource","code":"wikisource","sitename":"Wikizdroje"},{"url":"https://cs.wikiversity.org","dbname":"cswikiversity","code":"wikiversity","sitename":"Wikiverzita"}],"dir":"ltr","localname":"Tsjeggies"},"60":{"code":"csb","name":"kaszëbsczi","site":[{"url":"https://csb.wikipedia.org","dbname":"csbwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://csb.wiktionary.org","dbname":"csbwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Kashubian"},"61":{"code":"cu","name":"словѣньскъ / ⰔⰎⰑⰂⰡⰐⰠⰔⰍⰟ","site":[{"url":"https://cu.wikipedia.org","dbname":"cuwiki","code":"wiki","sitename":"Википєдїꙗ"}],"dir":"ltr","localname":"Kerkslawies"},"62":{"code":"cv","name":"чӑвашла","site":[{"url":"https://cv.wikipedia.org","dbname":"cvwiki","code":"wiki","sitename":"Википеди"},{"url":"https://cv.wikibooks.org","dbname":"cvwikibooks","code":"wikibooks","sitename":"Wikibooks"}],"dir":"ltr","localname":"Chuvash"},"63":{"code":"cy","name":"Cymraeg","site":[{"url":"https://cy.wikipedia.org","dbname":"cywiki","code":"wiki","sitename":"Wicipedia"},{"url":"https://cy.wiktionary.org","dbname":"cywiktionary","code":"wiktionary","sitename":"Wiciadur"},{"url":"https://cy.wikibooks.org","dbname":"cywikibooks","code":"wikibooks","sitename":"Wicilyfrau"},{"url":"https://cy.wikiquote.org","dbname":"cywikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://cy.wikisource.org","dbname":"cywikisource","code":"wikisource","sitename":"Wicidestun"}],"dir":"ltr","localname":"Wallies"},"64":{"code":"da","name":"dansk","site":[{"url":"https://da.wikipedia.org","dbname":"dawiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://da.wiktionary.org","dbname":"dawiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://da.wikibooks.org","dbname":"dawikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://da.wikiquote.org","dbname":"dawikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://da.wikisource.org","dbname":"dawikisource","code":"wikisource","sitename":"Wikisource"}],"dir":"ltr","localname":"Deens"},"65":{"code":"dag","name":"dagbanli","site":[{"url":"https://dag.wikipedia.org","dbname":"dagwiki","code":"wiki","sitename":"Dagbani Wikipedia"}],"dir":"ltr","localname":"Dagbani"},"66":{"code":"de","name":"Deutsch","site":[{"url":"https://de.wikipedia.org","dbname":"dewiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://de.wiktionary.org","dbname":"dewiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://de.wikibooks.org","dbname":"dewikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://de.wikinews.org","dbname":"dewikinews","code":"wikinews","sitename":"Wikinews"},{"url":"https://de.wikiquote.org","dbname":"dewikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://de.wikisource.org","dbname":"dewikisource","code":"wikisource","sitename":"Wikisource"},{"url":"https://de.wikiversity.org","dbname":"dewikiversity","code":"wikiversity","sitename":"Wikiversity"},{"url":"https://de.wikivoyage.org","dbname":"dewikivoyage","code":"wikivoyage","sitename":"Wikivoyage"}],"dir":"ltr","localname":"Duits"},"67":{"code":"din","name":"Thuɔŋjäŋ","site":[{"url":"https://din.wikipedia.org","dbname":"dinwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Dinka"},"68":{"code":"diq","name":"Zazaki","site":[{"url":"https://diq.wikipedia.org","dbname":"diqwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://diq.wiktionary.org","dbname":"diqwiktionary","code":"wiktionary","sitename":"Wikiqısebend"}],"dir":"ltr","localname":"Zazaki"},"69":{"code":"dsb","name":"dolnoserbski","site":[{"url":"https://dsb.wikipedia.org","dbname":"dsbwiki","code":"wiki","sitename":"Wikipedija"}],"dir":"ltr","localname":"Benedesorbies"},"70":{"code":"dty","name":"डोटेली","site":[{"url":"https://dty.wikipedia.org","dbname":"dtywiki","code":"wiki","sitename":"विकिपिडिया"}],"dir":"ltr","localname":"Doteli"},"71":{"code":"dv","name":"ދިވެހިބަސް","site":[{"url":"https://dv.wikipedia.org","dbname":"dvwiki","code":"wiki","sitename":"ވިކިޕީޑިއާ"},{"url":"https://dv.wiktionary.org","dbname":"dvwiktionary","code":"wiktionary","sitename":"ވިކިރަދީފު"}],"dir":"rtl","localname":"Divehi"},"72":{"code":"dz","name":"ཇོང་ཁ","site":[{"url":"https://dz.wikipedia.org","dbname":"dzwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://dz.wiktionary.org","dbname":"dzwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true}],"dir":"ltr","localname":"Dzongkha"},"73":{"code":"ee","name":"eʋegbe","site":[{"url":"https://ee.wikipedia.org","dbname":"eewiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Ewe"},"74":{"code":"el","name":"Ελληνικά","site":[{"url":"https://el.wikipedia.org","dbname":"elwiki","code":"wiki","sitename":"Βικιπαίδεια"},{"url":"https://el.wiktionary.org","dbname":"elwiktionary","code":"wiktionary","sitename":"Βικιλεξικό"},{"url":"https://el.wikibooks.org","dbname":"elwikibooks","code":"wikibooks","sitename":"Βικιβιβλία"},{"url":"https://el.wikinews.org","dbname":"elwikinews","code":"wikinews","sitename":"Βικινέα"},{"url":"https://el.wikiquote.org","dbname":"elwikiquote","code":"wikiquote","sitename":"Βικιφθέγματα"},{"url":"https://el.wikisource.org","dbname":"elwikisource","code":"wikisource","sitename":"Βικιθήκη"},{"url":"https://el.wikiversity.org","dbname":"elwikiversity","code":"wikiversity","sitename":"Βικιεπιστήμιο"},{"url":"https://el.wikivoyage.org","dbname":"elwikivoyage","code":"wikivoyage","sitename":"Βικιταξίδια"}],"dir":"ltr","localname":"Grieks"},"75":{"code":"eml","name":"emiliàn e rumagnòl","site":[{"url":"https://eml.wikipedia.org","dbname":"emlwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Emiliano-Romagnolo"},"76":{"code":"en","name":"English","site":[{"url":"https://en.wikipedia.org","dbname":"enwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://en.wiktionary.org","dbname":"enwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://en.wikibooks.org","dbname":"enwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://en.wikinews.org","dbname":"enwikinews","code":"wikinews","sitename":"Wikinews"},{"url":"https://en.wikiquote.org","dbname":"enwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://en.wikisource.org","dbname":"enwikisource","code":"wikisource","sitename":"Wikisource"},{"url":"https://en.wikiversity.org","dbname":"enwikiversity","code":"wikiversity","sitename":"Wikiversity"},{"url":"https://en.wikivoyage.org","dbname":"enwikivoyage","code":"wikivoyage","sitename":"Wikivoyage"}],"dir":"ltr","localname":"Engels"},"77":{"code":"eo","name":"Esperanto","site":[{"url":"https://eo.wikipedia.org","dbname":"eowiki","code":"wiki","sitename":"Vikipedio"},{"url":"https://eo.wiktionary.org","dbname":"eowiktionary","code":"wiktionary","sitename":"Vikivortaro"},{"url":"https://eo.wikibooks.org","dbname":"eowikibooks","code":"wikibooks","sitename":"Vikilibroj"},{"url":"https://eo.wikinews.org","dbname":"eowikinews","code":"wikinews","sitename":"Vikinovaĵoj"},{"url":"https://eo.wikiquote.org","dbname":"eowikiquote","code":"wikiquote","sitename":"Vikicitaro"},{"url":"https://eo.wikisource.org","dbname":"eowikisource","code":"wikisource","sitename":"Vikifontaro"},{"url":"https://eo.wikivoyage.org","dbname":"eowikivoyage","code":"wikivoyage","sitename":"Vikivojaĝo"}],"dir":"ltr","localname":"Esperanto"},"78":{"code":"es","name":"español","site":[{"url":"https://es.wikipedia.org","dbname":"eswiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://es.wiktionary.org","dbname":"eswiktionary","code":"wiktionary","sitename":"Wikcionario"},{"url":"https://es.wikibooks.org","dbname":"eswikibooks","code":"wikibooks","sitename":"Wikilibros"},{"url":"https://es.wikinews.org","dbname":"eswikinews","code":"wikinews","sitename":"Wikinoticias"},{"url":"https://es.wikiquote.org","dbname":"eswikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://es.wikisource.org","dbname":"eswikisource","code":"wikisource","sitename":"Wikisource"},{"url":"https://es.wikiversity.org","dbname":"eswikiversity","code":"wikiversity","sitename":"Wikiversidad"},{"url":"https://es.wikivoyage.org","dbname":"eswikivoyage","code":"wikivoyage","sitename":"Wikiviajes"}],"dir":"ltr","localname":"Spaans"},"79":{"code":"et","name":"eesti","site":[{"url":"https://et.wikipedia.org","dbname":"etwiki","code":"wiki","sitename":"Vikipeedia"},{"url":"https://et.wiktionary.org","dbname":"etwiktionary","code":"wiktionary","sitename":"Vikisõnastik"},{"url":"https://et.wikibooks.org","dbname":"etwikibooks","code":"wikibooks","sitename":"Vikiõpikud"},{"url":"https://et.wikiquote.org","dbname":"etwikiquote","code":"wikiquote","sitename":"Vikitsitaadid"},{"url":"https://et.wikisource.org","dbname":"etwikisource","code":"wikisource","sitename":"Vikitekstid"}],"dir":"ltr","localname":"Estnies"},"80":{"code":"eu","name":"euskara","site":[{"url":"https://eu.wikipedia.org","dbname":"euwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://eu.wiktionary.org","dbname":"euwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://eu.wikibooks.org","dbname":"euwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://eu.wikiquote.org","dbname":"euwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://eu.wikisource.org","dbname":"euwikisource","code":"wikisource","sitename":"Wikiteka"}],"dir":"ltr","localname":"Baskies"},"81":{"code":"ext","name":"estremeñu","site":[{"url":"https://ext.wikipedia.org","dbname":"extwiki","code":"wiki","sitename":"Güiquipeya"}],"dir":"ltr","localname":"Extremaduran"},"82":{"code":"fa","name":"فارسی","site":[{"url":"https://fa.wikipedia.org","dbname":"fawiki","code":"wiki","sitename":"ویکی‌پدیا"},{"url":"https://fa.wiktionary.org","dbname":"fawiktionary","code":"wiktionary","sitename":"ویکی‌واژه"},{"url":"https://fa.wikibooks.org","dbname":"fawikibooks","code":"wikibooks","sitename":"ویکی‌کتاب"},{"url":"https://fa.wikinews.org","dbname":"fawikinews","code":"wikinews","sitename":"ویکی‌خبر"},{"url":"https://fa.wikiquote.org","dbname":"fawikiquote","code":"wikiquote","sitename":"ویکی‌گفتاورد"},{"url":"https://fa.wikisource.org","dbname":"fawikisource","code":"wikisource","sitename":"ویکی‌نبشته"},{"url":"https://fa.wikivoyage.org","dbname":"fawikivoyage","code":"wikivoyage","sitename":"ویکی‌سفر"}],"dir":"rtl","localname":"Persies"},"83":{"code":"ff","name":"Fulfulde","site":[{"url":"https://ff.wikipedia.org","dbname":"ffwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Fulah"},"84":{"code":"fi","name":"suomi","site":[{"url":"https://fi.wikipedia.org","dbname":"fiwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://fi.wiktionary.org","dbname":"fiwiktionary","code":"wiktionary","sitename":"Wikisanakirja"},{"url":"https://fi.wikibooks.org","dbname":"fiwikibooks","code":"wikibooks","sitename":"Wikikirjasto"},{"url":"https://fi.wikinews.org","dbname":"fiwikinews","code":"wikinews","sitename":"Wikiuutiset"},{"url":"https://fi.wikiquote.org","dbname":"fiwikiquote","code":"wikiquote","sitename":"Wikisitaatit"},{"url":"https://fi.wikisource.org","dbname":"fiwikisource","code":"wikisource","sitename":"Wikiaineisto"},{"url":"https://fi.wikiversity.org","dbname":"fiwikiversity","code":"wikiversity","sitename":"Wikiopisto"},{"url":"https://fi.wikivoyage.org","dbname":"fiwikivoyage","code":"wikivoyage","sitename":"Wikimatkat"}],"dir":"ltr","localname":"Fins"},"85":{"code":"fiu-vro","name":"võro","site":[{"url":"https://fiu-vro.wikipedia.org","dbname":"fiu_vrowiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"võro"},"86":{"code":"fj","name":"Na Vosa Vakaviti","site":[{"url":"https://fj.wikipedia.org","dbname":"fjwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://fj.wiktionary.org","dbname":"fjwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Fidjiaans"},"87":{"code":"fo","name":"føroyskt","site":[{"url":"https://fo.wikipedia.org","dbname":"fowiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://fo.wiktionary.org","dbname":"fowiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://fo.wikisource.org","dbname":"fowikisource","code":"wikisource","sitename":"Wikiheimild"}],"dir":"ltr","localname":"Faroëes"},"88":{"code":"fr","name":"français","site":[{"url":"https://fr.wikipedia.org","dbname":"frwiki","code":"wiki","sitename":"Wikipédia"},{"url":"https://fr.wiktionary.org","dbname":"frwiktionary","code":"wiktionary","sitename":"Wiktionnaire"},{"url":"https://fr.wikibooks.org","dbname":"frwikibooks","code":"wikibooks","sitename":"Wikilivres"},{"url":"https://fr.wikinews.org","dbname":"frwikinews","code":"wikinews","sitename":"Wikinews"},{"url":"https://fr.wikiquote.org","dbname":"frwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://fr.wikisource.org","dbname":"frwikisource","code":"wikisource","sitename":"Wikisource"},{"url":"https://fr.wikiversity.org","dbname":"frwikiversity","code":"wikiversity","sitename":"Wikiversité"},{"url":"https://fr.wikivoyage.org","dbname":"frwikivoyage","code":"wikivoyage","sitename":"Wikivoyage"}],"dir":"ltr","localname":"Frans"},"89":{"code":"frp","name":"arpetan","site":[{"url":"https://frp.wikipedia.org","dbname":"frpwiki","code":"wiki","sitename":"Vouiquipèdia"}],"dir":"ltr","localname":"Arpitan"},"90":{"code":"frr","name":"Nordfriisk","site":[{"url":"https://frr.wikipedia.org","dbname":"frrwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Northern Frisian"},"91":{"code":"fur","name":"furlan","site":[{"url":"https://fur.wikipedia.org","dbname":"furwiki","code":"wiki","sitename":"Vichipedie"}],"dir":"ltr","localname":"Friuliaans"},"92":{"code":"fy","name":"Frysk","site":[{"url":"https://fy.wikipedia.org","dbname":"fywiki","code":"wiki","sitename":"Wikipedy"},{"url":"https://fy.wiktionary.org","dbname":"fywiktionary","code":"wiktionary","sitename":"Wikiwurdboek"},{"url":"https://fy.wikibooks.org","dbname":"fywikibooks","code":"wikibooks","sitename":"Wikibooks"}],"dir":"ltr","localname":"Fries"},"93":{"code":"ga","name":"Gaeilge","site":[{"url":"https://ga.wikipedia.org","dbname":"gawiki","code":"wiki","sitename":"Vicipéid"},{"url":"https://ga.wiktionary.org","dbname":"gawiktionary","code":"wiktionary","sitename":"Vicífhoclóir"},{"url":"https://ga.wikibooks.org","dbname":"gawikibooks","code":"wikibooks","sitename":"Vicíleabhair","closed":true},{"url":"https://ga.wikiquote.org","dbname":"gawikiquote","code":"wikiquote","sitename":"Vicísliocht","closed":true}],"dir":"ltr","localname":"Iers"},"94":{"code":"gag","name":"Gagauz","site":[{"url":"https://gag.wikipedia.org","dbname":"gagwiki","code":"wiki","sitename":"Vikipediya"}],"dir":"ltr","localname":"Gagauz"},"95":{"code":"gan","name":"贛語","site":[{"url":"https://gan.wikipedia.org","dbname":"ganwiki","code":"wiki","sitename":"維基百科"}],"dir":"ltr","localname":"Gan-Sjinees"},"96":{"code":"gcr","name":"kriyòl gwiyannen","site":[{"url":"https://gcr.wikipedia.org","dbname":"gcrwiki","code":"wiki","sitename":"Wikipédja"}],"dir":"ltr","localname":"Guianan Creole"},"97":{"code":"gd","name":"Gàidhlig","site":[{"url":"https://gd.wikipedia.org","dbname":"gdwiki","code":"wiki","sitename":"Uicipeid"},{"url":"https://gd.wiktionary.org","dbname":"gdwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Skotse Gallies"},"98":{"code":"gl","name":"galego","site":[{"url":"https://gl.wikipedia.org","dbname":"glwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://gl.wiktionary.org","dbname":"glwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://gl.wikibooks.org","dbname":"glwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://gl.wikiquote.org","dbname":"glwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://gl.wikisource.org","dbname":"glwikisource","code":"wikisource","sitename":"Wikisource"}],"dir":"ltr","localname":"Galisies"},"99":{"code":"glk","name":"گیلکی","site":[{"url":"https://glk.wikipedia.org","dbname":"glkwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"rtl","localname":"Gilaki"},"100":{"code":"gn","name":"Avañe'ẽ","site":[{"url":"https://gn.wikipedia.org","dbname":"gnwiki","code":"wiki","sitename":"Vikipetã"},{"url":"https://gn.wiktionary.org","dbname":"gnwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://gn.wikibooks.org","dbname":"gnwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Guarani"},"101":{"code":"gom","name":"गोंयची कोंकणी / Gõychi Konknni","site":[{"url":"https://gom.wikipedia.org","dbname":"gomwiki","code":"wiki","sitename":"विकिपीडिया"},{"url":"https://gom.wiktionary.org","dbname":"gomwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Goan Konkani"},"102":{"code":"gor","name":"Bahasa Hulontalo","site":[{"url":"https://gor.wikipedia.org","dbname":"gorwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://gor.wiktionary.org","dbname":"gorwiktionary","code":"wiktionary","sitename":"Wikikamus"}],"dir":"ltr","localname":"Gorontalo"},"103":{"code":"got","name":"𐌲𐌿𐍄𐌹𐍃𐌺","site":[{"url":"https://got.wikipedia.org","dbname":"gotwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://got.wikibooks.org","dbname":"gotwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Goties"},"104":{"code":"gu","name":"ગુજરાતી","site":[{"url":"https://gu.wikipedia.org","dbname":"guwiki","code":"wiki","sitename":"વિકિપીડિયા"},{"url":"https://gu.wiktionary.org","dbname":"guwiktionary","code":"wiktionary","sitename":"વિકિકોશ"},{"url":"https://gu.wikibooks.org","dbname":"guwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://gu.wikiquote.org","dbname":"guwikiquote","code":"wikiquote","sitename":"વિકિસૂક્તિ"},{"url":"https://gu.wikisource.org","dbname":"guwikisource","code":"wikisource","sitename":"વિકિસ્રોત"}],"dir":"ltr","localname":"Goedjarati"},"105":{"code":"guw","name":"gungbe","site":[{"url":"https://guw.wikipedia.org","dbname":"guwwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://guw.wiktionary.org","dbname":"guwwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://guw.wikiquote.org","dbname":"guwwikiquote","code":"wikiquote","sitename":"Wikiquote"}],"dir":"ltr","localname":"Gun"},"106":{"code":"gv","name":"Gaelg","site":[{"url":"https://gv.wikipedia.org","dbname":"gvwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://gv.wiktionary.org","dbname":"gvwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Manx"},"107":{"code":"ha","name":"Hausa","site":[{"url":"https://ha.wikipedia.org","dbname":"hawiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ha.wiktionary.org","dbname":"hawiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Hausa"},"108":{"code":"hak","name":"客家語/Hak-kâ-ngî","site":[{"url":"https://hak.wikipedia.org","dbname":"hakwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Hakka-Sjinees"},"109":{"code":"haw","name":"Hawaiʻi","site":[{"url":"https://haw.wikipedia.org","dbname":"hawwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Hawais"},"110":{"code":"he","name":"עברית","site":[{"url":"https://he.wikipedia.org","dbname":"hewiki","code":"wiki","sitename":"ויקיפדיה"},{"url":"https://he.wiktionary.org","dbname":"hewiktionary","code":"wiktionary","sitename":"ויקימילון"},{"url":"https://he.wikibooks.org","dbname":"hewikibooks","code":"wikibooks","sitename":"ויקיספר"},{"url":"https://he.wikinews.org","dbname":"hewikinews","code":"wikinews","sitename":"ויקיחדשות"},{"url":"https://he.wikiquote.org","dbname":"hewikiquote","code":"wikiquote","sitename":"ויקיציטוט"},{"url":"https://he.wikisource.org","dbname":"hewikisource","code":"wikisource","sitename":"ויקיטקסט"},{"url":"https://he.wikivoyage.org","dbname":"hewikivoyage","code":"wikivoyage","sitename":"ויקימסע"}],"dir":"rtl","localname":"Hebreeus"},"111":{"code":"hi","name":"हिन्दी","site":[{"url":"https://hi.wikipedia.org","dbname":"hiwiki","code":"wiki","sitename":"विकिपीडिया"},{"url":"https://hi.wiktionary.org","dbname":"hiwiktionary","code":"wiktionary","sitename":"विक्षनरी"},{"url":"https://hi.wikibooks.org","dbname":"hiwikibooks","code":"wikibooks","sitename":"विकिपुस्तक"},{"url":"https://hi.wikiquote.org","dbname":"hiwikiquote","code":"wikiquote","sitename":"विकिसूक्ति"},{"url":"https://hi.wikisource.org","dbname":"hiwikisource","code":"wikisource","sitename":"विकिस्रोत"},{"url":"https://hi.wikiversity.org","dbname":"hiwikiversity","code":"wikiversity","sitename":"विकिविश्वविद्यालय"},{"url":"https://hi.wikivoyage.org","dbname":"hiwikivoyage","code":"wikivoyage","sitename":"विकियात्रा"}],"dir":"ltr","localname":"Hindi"},"112":{"code":"hif","name":"Fiji Hindi","site":[{"url":"https://hif.wikipedia.org","dbname":"hifwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://hif.wiktionary.org","dbname":"hifwiktionary","code":"wiktionary","sitename":"Sabdkosh"}],"dir":"ltr","localname":"Fiji Hindi"},"113":{"code":"ho","name":"Hiri Motu","site":[{"url":"https://ho.wikipedia.org","dbname":"howiki","code":"wiki","sitename":"Wikipedia","closed":true}],"dir":"ltr","localname":"Hiri Motu"},"114":{"code":"hr","name":"hrvatski","site":[{"url":"https://hr.wikipedia.org","dbname":"hrwiki","code":"wiki","sitename":"Wikipedija"},{"url":"https://hr.wiktionary.org","dbname":"hrwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://hr.wikibooks.org","dbname":"hrwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://hr.wikiquote.org","dbname":"hrwikiquote","code":"wikiquote","sitename":"Wikicitat"},{"url":"https://hr.wikisource.org","dbname":"hrwikisource","code":"wikisource","sitename":"Wikizvor"}],"dir":"ltr","localname":"Kroaties"},"115":{"code":"hsb","name":"hornjoserbsce","site":[{"url":"https://hsb.wikipedia.org","dbname":"hsbwiki","code":"wiki","sitename":"Wikipedija"},{"url":"https://hsb.wiktionary.org","dbname":"hsbwiktionary","code":"wiktionary","sitename":"Wikisłownik"}],"dir":"ltr","localname":"Oppersorbies"},"116":{"code":"ht","name":"Kreyòl ayisyen","site":[{"url":"https://ht.wikipedia.org","dbname":"htwiki","code":"wiki","sitename":"Wikipedya"},{"url":"https://ht.wikisource.org","dbname":"htwikisource","code":"wikisource","sitename":"Wikisòrs","closed":true}],"dir":"ltr","localname":"Haïtiaans"},"117":{"code":"hu","name":"magyar","site":[{"url":"https://hu.wikipedia.org","dbname":"huwiki","code":"wiki","sitename":"Wikipédia"},{"url":"https://hu.wiktionary.org","dbname":"huwiktionary","code":"wiktionary","sitename":"Wikiszótár"},{"url":"https://hu.wikibooks.org","dbname":"huwikibooks","code":"wikibooks","sitename":"Wikikönyvek"},{"url":"https://hu.wikinews.org","dbname":"huwikinews","code":"wikinews","sitename":"Wikihírek","closed":true},{"url":"https://hu.wikiquote.org","dbname":"huwikiquote","code":"wikiquote","sitename":"Wikidézet"},{"url":"https://hu.wikisource.org","dbname":"huwikisource","code":"wikisource","sitename":"Wikiforrás"}],"dir":"ltr","localname":"Hongaars"},"118":{"code":"hy","name":"հայերեն","site":[{"url":"https://hy.wikipedia.org","dbname":"hywiki","code":"wiki","sitename":"Վիքիպեդիա"},{"url":"https://hy.wiktionary.org","dbname":"hywiktionary","code":"wiktionary","sitename":"Վիքիբառարան"},{"url":"https://hy.wikibooks.org","dbname":"hywikibooks","code":"wikibooks","sitename":"Վիքիգրքեր"},{"url":"https://hy.wikiquote.org","dbname":"hywikiquote","code":"wikiquote","sitename":"Վիքիքաղվածք"},{"url":"https://hy.wikisource.org","dbname":"hywikisource","code":"wikisource","sitename":"Վիքիդարան"}],"dir":"ltr","localname":"Armeens"},"119":{"code":"hyw","name":"Արեւմտահայերէն","site":[{"url":"https://hyw.wikipedia.org","dbname":"hywwiki","code":"wiki","sitename":"Ուիքիփետիա"}],"dir":"ltr","localname":"Western Armenian"},"120":{"code":"hz","name":"Otsiherero","site":[{"url":"https://hz.wikipedia.org","dbname":"hzwiki","code":"wiki","sitename":"Wikipedia","closed":true}],"dir":"ltr","localname":"Herero"},"121":{"code":"ia","name":"interlingua","site":[{"url":"https://ia.wikipedia.org","dbname":"iawiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ia.wiktionary.org","dbname":"iawiktionary","code":"wiktionary","sitename":"Wiktionario"},{"url":"https://ia.wikibooks.org","dbname":"iawikibooks","code":"wikibooks","sitename":"Wikibooks"}],"dir":"ltr","localname":"Interlingua"},"122":{"code":"id","name":"Bahasa Indonesia","site":[{"url":"https://id.wikipedia.org","dbname":"idwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://id.wiktionary.org","dbname":"idwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://id.wikibooks.org","dbname":"idwikibooks","code":"wikibooks","sitename":"Wikibuku"},{"url":"https://id.wikiquote.org","dbname":"idwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://id.wikisource.org","dbname":"idwikisource","code":"wikisource","sitename":"Wikisource"}],"dir":"ltr","localname":"Indonesies"},"123":{"code":"ie","name":"Interlingue","site":[{"url":"https://ie.wikipedia.org","dbname":"iewiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ie.wiktionary.org","dbname":"iewiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://ie.wikibooks.org","dbname":"iewikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Interlingue"},"124":{"code":"ig","name":"Igbo","site":[{"url":"https://ig.wikipedia.org","dbname":"igwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ig.wiktionary.org","dbname":"igwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://ig.wikiquote.org","dbname":"igwikiquote","code":"wikiquote","sitename":"Wikiquote"}],"dir":"ltr","localname":"Igbo"},"125":{"code":"ii","name":"ꆇꉙ","site":[{"url":"https://ii.wikipedia.org","dbname":"iiwiki","code":"wiki","sitename":"Wikipedia","closed":true}],"dir":"ltr","localname":"Sichuan Yi"},"126":{"code":"ik","name":"Iñupiatun","site":[{"url":"https://ik.wikipedia.org","dbname":"ikwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ik.wiktionary.org","dbname":"ikwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true}],"dir":"ltr","localname":"Inupiaq"},"127":{"code":"ilo","name":"Ilokano","site":[{"url":"https://ilo.wikipedia.org","dbname":"ilowiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Iloko"},"128":{"code":"inh","name":"гӀалгӀай","site":[{"url":"https://inh.wikipedia.org","dbname":"inhwiki","code":"wiki","sitename":"Википеди"}],"dir":"ltr","localname":"Ingush"},"129":{"code":"io","name":"Ido","site":[{"url":"https://io.wikipedia.org","dbname":"iowiki","code":"wiki","sitename":"Wikipedio"},{"url":"https://io.wiktionary.org","dbname":"iowiktionary","code":"wiktionary","sitename":"Wikivortaro"}],"dir":"ltr","localname":"Ido"},"130":{"code":"is","name":"íslenska","site":[{"url":"https://is.wikipedia.org","dbname":"iswiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://is.wiktionary.org","dbname":"iswiktionary","code":"wiktionary","sitename":"Wikiorðabók"},{"url":"https://is.wikibooks.org","dbname":"iswikibooks","code":"wikibooks","sitename":"Wikibækur"},{"url":"https://is.wikiquote.org","dbname":"iswikiquote","code":"wikiquote","sitename":"Wikivitnun"},{"url":"https://is.wikisource.org","dbname":"iswikisource","code":"wikisource","sitename":"Wikiheimild"}],"dir":"ltr","localname":"Yslands"},"131":{"code":"it","name":"italiano","site":[{"url":"https://it.wikipedia.org","dbname":"itwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://it.wiktionary.org","dbname":"itwiktionary","code":"wiktionary","sitename":"Wikizionario"},{"url":"https://it.wikibooks.org","dbname":"itwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://it.wikinews.org","dbname":"itwikinews","code":"wikinews","sitename":"Wikinotizie"},{"url":"https://it.wikiquote.org","dbname":"itwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://it.wikisource.org","dbname":"itwikisource","code":"wikisource","sitename":"Wikisource"},{"url":"https://it.wikiversity.org","dbname":"itwikiversity","code":"wikiversity","sitename":"Wikiversità"},{"url":"https://it.wikivoyage.org","dbname":"itwikivoyage","code":"wikivoyage","sitename":"Wikivoyage"}],"dir":"ltr","localname":"Italiaans"},"132":{"code":"iu","name":"ᐃᓄᒃᑎᑐᑦ/inuktitut","site":[{"url":"https://iu.wikipedia.org","dbname":"iuwiki","code":"wiki","sitename":"ᐅᐃᑭᐱᑎᐊ"},{"url":"https://iu.wiktionary.org","dbname":"iuwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Inoektitoet"},"133":{"code":"ja","name":"日本語","site":[{"url":"https://ja.wikipedia.org","dbname":"jawiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ja.wiktionary.org","dbname":"jawiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://ja.wikibooks.org","dbname":"jawikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://ja.wikinews.org","dbname":"jawikinews","code":"wikinews","sitename":"ウィキニュース"},{"url":"https://ja.wikiquote.org","dbname":"jawikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://ja.wikisource.org","dbname":"jawikisource","code":"wikisource","sitename":"Wikisource"},{"url":"https://ja.wikiversity.org","dbname":"jawikiversity","code":"wikiversity","sitename":"ウィキバーシティ"},{"url":"https://ja.wikivoyage.org","dbname":"jawikivoyage","code":"wikivoyage","sitename":"ウィキボヤージュ"}],"dir":"ltr","localname":"Japannees"},"134":{"code":"jam","name":"Patois","site":[{"url":"https://jam.wikipedia.org","dbname":"jamwiki","code":"wiki","sitename":"Wikipidia"}],"dir":"ltr","localname":"Jamaican Creole English"},"135":{"code":"jbo","name":"la .lojban.","site":[{"url":"https://jbo.wikipedia.org","dbname":"jbowiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://jbo.wiktionary.org","dbname":"jbowiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Lojban"},"136":{"code":"jv","name":"Jawa","site":[{"url":"https://jv.wikipedia.org","dbname":"jvwiki","code":"wiki","sitename":"Wikipédia"},{"url":"https://jv.wiktionary.org","dbname":"jvwiktionary","code":"wiktionary","sitename":"Wikisastra"},{"url":"https://jv.wikisource.org","dbname":"jvwikisource","code":"wikisource","sitename":"Wikisumber"}],"dir":"ltr","localname":"Javaans"},"137":{"code":"ka","name":"ქართული","site":[{"url":"https://ka.wikipedia.org","dbname":"kawiki","code":"wiki","sitename":"ვიკიპედია"},{"url":"https://ka.wiktionary.org","dbname":"kawiktionary","code":"wiktionary","sitename":"ვიქსიკონი"},{"url":"https://ka.wikibooks.org","dbname":"kawikibooks","code":"wikibooks","sitename":"ვიკიწიგნები"},{"url":"https://ka.wikiquote.org","dbname":"kawikiquote","code":"wikiquote","sitename":"ვიკიციტატა"}],"dir":"ltr","localname":"Georgies"},"138":{"code":"kaa","name":"Qaraqalpaqsha","site":[{"url":"https://kaa.wikipedia.org","dbname":"kaawiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Kara-Kalpak"},"139":{"code":"kab","name":"Taqbaylit","site":[{"url":"https://kab.wikipedia.org","dbname":"kabwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Kabyle"},"140":{"code":"kbd","name":"адыгэбзэ","site":[{"url":"https://kbd.wikipedia.org","dbname":"kbdwiki","code":"wiki","sitename":"Уикипедиэ"}],"dir":"ltr","localname":"Kabardiaans"},"141":{"code":"kbp","name":"Kabɩyɛ","site":[{"url":"https://kbp.wikipedia.org","dbname":"kbpwiki","code":"wiki","sitename":"Wikipediya"}],"dir":"ltr","localname":"Kabiye"},"142":{"code":"kcg","name":"Tyap","site":[{"url":"https://kcg.wikipedia.org","dbname":"kcgwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Tyap"},"143":{"code":"kg","name":"Kongo","site":[{"url":"https://kg.wikipedia.org","dbname":"kgwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Kongolees"},"144":{"code":"ki","name":"Gĩkũyũ","site":[{"url":"https://ki.wikipedia.org","dbname":"kiwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Kikuyu"},"145":{"code":"kj","name":"Kwanyama","site":[{"url":"https://kj.wikipedia.org","dbname":"kjwiki","code":"wiki","sitename":"Wikipedia","closed":true}],"dir":"ltr","localname":"Kuanyama"},"146":{"code":"kk","name":"қазақша","site":[{"url":"https://kk.wikipedia.org","dbname":"kkwiki","code":"wiki","sitename":"Уикипедия"},{"url":"https://kk.wiktionary.org","dbname":"kkwiktionary","code":"wiktionary","sitename":"Уикисөздік"},{"url":"https://kk.wikibooks.org","dbname":"kkwikibooks","code":"wikibooks","sitename":"Уикикітап"},{"url":"https://kk.wikiquote.org","dbname":"kkwikiquote","code":"wikiquote","sitename":"Уикидәйек","closed":true}],"dir":"ltr","localname":"Kazaks"},"147":{"code":"kl","name":"kalaallisut","site":[{"url":"https://kl.wikipedia.org","dbname":"klwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://kl.wiktionary.org","dbname":"klwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Kalaallisut"},"148":{"code":"km","name":"ភាសាខ្មែរ","site":[{"url":"https://km.wikipedia.org","dbname":"kmwiki","code":"wiki","sitename":"វិគីភីឌា"},{"url":"https://km.wiktionary.org","dbname":"kmwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://km.wikibooks.org","dbname":"kmwikibooks","code":"wikibooks","sitename":"Wikibooks"}],"dir":"ltr","localname":"Khmer"},"149":{"code":"kn","name":"ಕನ್ನಡ","site":[{"url":"https://kn.wikipedia.org","dbname":"knwiki","code":"wiki","sitename":"ವಿಕಿಪೀಡಿಯ"},{"url":"https://kn.wiktionary.org","dbname":"knwiktionary","code":"wiktionary","sitename":"ವಿಕ್ಷನರಿ"},{"url":"https://kn.wikibooks.org","dbname":"knwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://kn.wikiquote.org","dbname":"knwikiquote","code":"wikiquote","sitename":"ವಿಕಿಕೋಟ್"},{"url":"https://kn.wikisource.org","dbname":"knwikisource","code":"wikisource","sitename":"ವಿಕಿಸೋರ್ಸ್"}],"dir":"ltr","localname":"Kannada"},"150":{"code":"ko","name":"한국어","site":[{"url":"https://ko.wikipedia.org","dbname":"kowiki","code":"wiki","sitename":"위키백과"},{"url":"https://ko.wiktionary.org","dbname":"kowiktionary","code":"wiktionary","sitename":"위키낱말사전"},{"url":"https://ko.wikibooks.org","dbname":"kowikibooks","code":"wikibooks","sitename":"위키책"},{"url":"https://ko.wikinews.org","dbname":"kowikinews","code":"wikinews","sitename":"위키뉴스"},{"url":"https://ko.wikiquote.org","dbname":"kowikiquote","code":"wikiquote","sitename":"위키인용집"},{"url":"https://ko.wikisource.org","dbname":"kowikisource","code":"wikisource","sitename":"위키문헌"},{"url":"https://ko.wikiversity.org","dbname":"kowikiversity","code":"wikiversity","sitename":"위키배움터"}],"dir":"ltr","localname":"Koreaans"},"151":{"code":"koi","name":"перем коми","site":[{"url":"https://koi.wikipedia.org","dbname":"koiwiki","code":"wiki","sitename":"Википедия"}],"dir":"ltr","localname":"Komi-Permyaks"},"152":{"code":"kr","name":"kanuri","site":[{"url":"https://kr.wikipedia.org","dbname":"krwiki","code":"wiki","sitename":"Wikipedia","closed":true},{"url":"https://kr.wikiquote.org","dbname":"krwikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Kanuri"},"153":{"code":"krc","name":"къарачай-малкъар","site":[{"url":"https://krc.wikipedia.org","dbname":"krcwiki","code":"wiki","sitename":"Википедия"}],"dir":"ltr","localname":"Karachay-Balkar"},"154":{"code":"ks","name":"कॉशुर / کٲشُر","site":[{"url":"https://ks.wikipedia.org","dbname":"kswiki","code":"wiki","sitename":"وِکیٖپیٖڈیا"},{"url":"https://ks.wiktionary.org","dbname":"kswiktionary","code":"wiktionary","sitename":"وِکیٖلۄغَتھ"},{"url":"https://ks.wikibooks.org","dbname":"kswikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://ks.wikiquote.org","dbname":"kswikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"rtl","localname":"Kasjmirs"},"155":{"code":"ksh","name":"Ripoarisch","site":[{"url":"https://ksh.wikipedia.org","dbname":"kshwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Keuls"},"156":{"code":"ku","name":"kurdî","site":[{"url":"https://ku.wikipedia.org","dbname":"kuwiki","code":"wiki","sitename":"Wîkîpediya"},{"url":"https://ku.wiktionary.org","dbname":"kuwiktionary","code":"wiktionary","sitename":"Wîkîferheng"},{"url":"https://ku.wikibooks.org","dbname":"kuwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://ku.wikiquote.org","dbname":"kuwikiquote","code":"wikiquote","sitename":"Wikiquote"}],"dir":"ltr","localname":"Koerdies"},"157":{"code":"kv","name":"коми","site":[{"url":"https://kv.wikipedia.org","dbname":"kvwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Komi"},"158":{"code":"kw","name":"kernowek","site":[{"url":"https://kw.wikipedia.org","dbname":"kwwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://kw.wiktionary.org","dbname":"kwwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://kw.wikiquote.org","dbname":"kwwikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Kornies"},"159":{"code":"ky","name":"кыргызча","site":[{"url":"https://ky.wikipedia.org","dbname":"kywiki","code":"wiki","sitename":"Википедия"},{"url":"https://ky.wiktionary.org","dbname":"kywiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://ky.wikibooks.org","dbname":"kywikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://ky.wikiquote.org","dbname":"kywikiquote","code":"wikiquote","sitename":"Wikiquote"}],"dir":"ltr","localname":"Kirgisies"},"160":{"code":"la","name":"Latina","site":[{"url":"https://la.wikipedia.org","dbname":"lawiki","code":"wiki","sitename":"Vicipaedia"},{"url":"https://la.wiktionary.org","dbname":"lawiktionary","code":"wiktionary","sitename":"Victionarium"},{"url":"https://la.wikibooks.org","dbname":"lawikibooks","code":"wikibooks","sitename":"Vicilibri"},{"url":"https://la.wikiquote.org","dbname":"lawikiquote","code":"wikiquote","sitename":"Vicicitatio"},{"url":"https://la.wikisource.org","dbname":"lawikisource","code":"wikisource","sitename":"Wikisource"}],"dir":"ltr","localname":"Latyn"},"161":{"code":"lad","name":"Ladino","site":[{"url":"https://lad.wikipedia.org","dbname":"ladwiki","code":"wiki","sitename":"Vikipedya"}],"dir":"ltr","localname":"Ladino"},"162":{"code":"lb","name":"Lëtzebuergesch","site":[{"url":"https://lb.wikipedia.org","dbname":"lbwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://lb.wiktionary.org","dbname":"lbwiktionary","code":"wiktionary","sitename":"Wiktionnaire"},{"url":"https://lb.wikibooks.org","dbname":"lbwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://lb.wikiquote.org","dbname":"lbwikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Luxemburgs"},"163":{"code":"lbe","name":"лакку","site":[{"url":"https://lbe.wikipedia.org","dbname":"lbewiki","code":"wiki","sitename":"Википедия"}],"dir":"ltr","localname":"Lak"},"164":{"code":"lez","name":"лезги","site":[{"url":"https://lez.wikipedia.org","dbname":"lezwiki","code":"wiki","sitename":"Википедия"}],"dir":"ltr","localname":"Lezghies"},"165":{"code":"lfn","name":"Lingua Franca Nova","site":[{"url":"https://lfn.wikipedia.org","dbname":"lfnwiki","code":"wiki","sitename":"Vicipedia"}],"dir":"ltr","localname":"Lingua Franca Nova"},"166":{"code":"lg","name":"Luganda","site":[{"url":"https://lg.wikipedia.org","dbname":"lgwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Ganda"},"167":{"code":"li","name":"Limburgs","site":[{"url":"https://li.wikipedia.org","dbname":"liwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://li.wiktionary.org","dbname":"liwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://li.wikibooks.org","dbname":"liwikibooks","code":"wikibooks","sitename":"Wikibeuk"},{"url":"https://li.wikinews.org","dbname":"liwikinews","code":"wikinews","sitename":"Wikinuujs"},{"url":"https://li.wikiquote.org","dbname":"liwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://li.wikisource.org","dbname":"liwikisource","code":"wikisource","sitename":"Wikibrónne"}],"dir":"ltr","localname":"Limburgs"},"168":{"code":"lij","name":"Ligure","site":[{"url":"https://lij.wikipedia.org","dbname":"lijwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://lij.wikisource.org","dbname":"lijwikisource","code":"wikisource","sitename":"Wikivivàgna"}],"dir":"ltr","localname":"Ligurian"},"169":{"code":"lld","name":"Ladin","site":[{"url":"https://lld.wikipedia.org","dbname":"lldwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Ladin"},"170":{"code":"lmo","name":"lombard","site":[{"url":"https://lmo.wikipedia.org","dbname":"lmowiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://lmo.wiktionary.org","dbname":"lmowiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Lombard"},"171":{"code":"ln","name":"lingála","site":[{"url":"https://ln.wikipedia.org","dbname":"lnwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ln.wiktionary.org","dbname":"lnwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://ln.wikibooks.org","dbname":"lnwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Lingaals"},"172":{"code":"lo","name":"ລາວ","site":[{"url":"https://lo.wikipedia.org","dbname":"lowiki","code":"wiki","sitename":"ວິກິພີເດຍ"},{"url":"https://lo.wiktionary.org","dbname":"lowiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Lao"},"173":{"code":"lrc","name":"لۊری شومالی","site":[{"url":"https://lrc.wikipedia.org","dbname":"lrcwiki","code":"wiki","sitename":"ڤیکیپئدیا","closed":true}],"dir":"rtl","localname":"Noord-Luri"},"174":{"code":"lt","name":"lietuvių","site":[{"url":"https://lt.wikipedia.org","dbname":"ltwiki","code":"wiki","sitename":"Vikipedija"},{"url":"https://lt.wiktionary.org","dbname":"ltwiktionary","code":"wiktionary","sitename":"Vikižodynas"},{"url":"https://lt.wikibooks.org","dbname":"ltwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://lt.wikiquote.org","dbname":"ltwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://lt.wikisource.org","dbname":"ltwikisource","code":"wikisource","sitename":"Vikišaltiniai"}],"dir":"ltr","localname":"Litaus"},"175":{"code":"ltg","name":"latgaļu","site":[{"url":"https://ltg.wikipedia.org","dbname":"ltgwiki","code":"wiki","sitename":"Vikipedeja"}],"dir":"ltr","localname":"Latgalian"},"176":{"code":"lv","name":"latviešu","site":[{"url":"https://lv.wikipedia.org","dbname":"lvwiki","code":"wiki","sitename":"Vikipēdija"},{"url":"https://lv.wiktionary.org","dbname":"lvwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://lv.wikibooks.org","dbname":"lvwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Letties"},"177":{"code":"mad","name":"Madhurâ","site":[{"url":"https://mad.wikipedia.org","dbname":"madwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Madurees"},"178":{"code":"mai","name":"मैथिली","site":[{"url":"https://mai.wikipedia.org","dbname":"maiwiki","code":"wiki","sitename":"विकिपिडिया"}],"dir":"ltr","localname":"Maithili"},"179":{"code":"map-bms","name":"Basa Banyumasan","site":[{"url":"https://map-bms.wikipedia.org","dbname":"map_bmswiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Basa Banyumasan"},"180":{"code":"mdf","name":"мокшень","site":[{"url":"https://mdf.wikipedia.org","dbname":"mdfwiki","code":"wiki","sitename":"Википедиесь"}],"dir":"ltr","localname":"Moksha"},"181":{"code":"mg","name":"Malagasy","site":[{"url":"https://mg.wikipedia.org","dbname":"mgwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://mg.wiktionary.org","dbname":"mgwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://mg.wikibooks.org","dbname":"mgwikibooks","code":"wikibooks","sitename":"Wikibooks"}],"dir":"ltr","localname":"Malgassies"},"182":{"code":"mh","name":"Ebon","site":[{"url":"https://mh.wikipedia.org","dbname":"mhwiki","code":"wiki","sitename":"Wikipedia","closed":true},{"url":"https://mh.wiktionary.org","dbname":"mhwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true}],"dir":"ltr","localname":"Marshallees"},"183":{"code":"mhr","name":"олык марий","site":[{"url":"https://mhr.wikipedia.org","dbname":"mhrwiki","code":"wiki","sitename":"Википедий"}],"dir":"ltr","localname":"Eastern Mari"},"184":{"code":"mi","name":"Māori","site":[{"url":"https://mi.wikipedia.org","dbname":"miwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://mi.wiktionary.org","dbname":"miwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://mi.wikibooks.org","dbname":"miwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Maori"},"185":{"code":"min","name":"Minangkabau","site":[{"url":"https://min.wikipedia.org","dbname":"minwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://min.wiktionary.org","dbname":"minwiktionary","code":"wiktionary","sitename":"Wikikato"}],"dir":"ltr","localname":"Minangkabaus"},"186":{"code":"mk","name":"македонски","site":[{"url":"https://mk.wikipedia.org","dbname":"mkwiki","code":"wiki","sitename":"Википедија"},{"url":"https://mk.wiktionary.org","dbname":"mkwiktionary","code":"wiktionary","sitename":"Викиречник"},{"url":"https://mk.wikibooks.org","dbname":"mkwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://mk.wikisource.org","dbname":"mkwikisource","code":"wikisource","sitename":"Wikisource"}],"dir":"ltr","localname":"Masedonies"},"187":{"code":"ml","name":"മലയാളം","site":[{"url":"https://ml.wikipedia.org","dbname":"mlwiki","code":"wiki","sitename":"വിക്കിപീഡിയ"},{"url":"https://ml.wiktionary.org","dbname":"mlwiktionary","code":"wiktionary","sitename":"വിക്കിനിഘണ്ടു"},{"url":"https://ml.wikibooks.org","dbname":"mlwikibooks","code":"wikibooks","sitename":"വിക്കിപാഠശാല"},{"url":"https://ml.wikiquote.org","dbname":"mlwikiquote","code":"wikiquote","sitename":"വിക്കിചൊല്ലുകൾ"},{"url":"https://ml.wikisource.org","dbname":"mlwikisource","code":"wikisource","sitename":"വിക്കിഗ്രന്ഥശാല"}],"dir":"ltr","localname":"Malabaars"},"188":{"code":"mn","name":"монгол","site":[{"url":"https://mn.wikipedia.org","dbname":"mnwiki","code":"wiki","sitename":"Википедиа"},{"url":"https://mn.wiktionary.org","dbname":"mnwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://mn.wikibooks.org","dbname":"mnwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Mongools"},"189":{"code":"mni","name":"ꯃꯤꯇꯩ ꯂꯣꯟ","site":[{"url":"https://mni.wikipedia.org","dbname":"mniwiki","code":"wiki","sitename":"ꯋꯤꯀꯤꯄꯦꯗꯤꯌꯥ"},{"url":"https://mni.wiktionary.org","dbname":"mniwiktionary","code":"wiktionary","sitename":"ꯋꯤꯛꯁꯟꯅꯔꯤ"}],"dir":"ltr","localname":"Manipuri"},"190":{"code":"mnw","name":"ဘာသာ မန်","site":[{"url":"https://mnw.wikipedia.org","dbname":"mnwwiki","code":"wiki","sitename":"ဝဳကဳပဳဒဳယာ"},{"url":"https://mnw.wiktionary.org","dbname":"mnwwiktionary","code":"wiktionary","sitename":"ဝိက်ရှေန်နရဳ"}],"dir":"ltr","localname":"Mon"},"191":{"code":"mo","name":"молдовеняскэ","site":[],"dir":"ltr","localname":"Moldovan"},"192":{"code":"mr","name":"मराठी","site":[{"url":"https://mr.wikipedia.org","dbname":"mrwiki","code":"wiki","sitename":"विकिपीडिया"},{"url":"https://mr.wiktionary.org","dbname":"mrwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://mr.wikibooks.org","dbname":"mrwikibooks","code":"wikibooks","sitename":"विकिबुक्स"},{"url":"https://mr.wikiquote.org","dbname":"mrwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://mr.wikisource.org","dbname":"mrwikisource","code":"wikisource","sitename":"विकिस्रोत"}],"dir":"ltr","localname":"Marathi"},"193":{"code":"mrj","name":"кырык мары","site":[{"url":"https://mrj.wikipedia.org","dbname":"mrjwiki","code":"wiki","sitename":"Википеди"}],"dir":"ltr","localname":"Western Mari"},"194":{"code":"ms","name":"Bahasa Melayu","site":[{"url":"https://ms.wikipedia.org","dbname":"mswiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ms.wiktionary.org","dbname":"mswiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://ms.wikibooks.org","dbname":"mswikibooks","code":"wikibooks","sitename":"Wikibooks"}],"dir":"ltr","localname":"Maleis"},"195":{"code":"mt","name":"Malti","site":[{"url":"https://mt.wikipedia.org","dbname":"mtwiki","code":"wiki","sitename":"Wikipedija"},{"url":"https://mt.wiktionary.org","dbname":"mtwiktionary","code":"wiktionary","sitename":"Wikizzjunarju"}],"dir":"ltr","localname":"Maltees"},"196":{"code":"mus","name":"Mvskoke","site":[{"url":"https://mus.wikipedia.org","dbname":"muswiki","code":"wiki","sitename":"Wikipedia","closed":true}],"dir":"ltr","localname":"Kreek"},"197":{"code":"mwl","name":"Mirandés","site":[{"url":"https://mwl.wikipedia.org","dbname":"mwlwiki","code":"wiki","sitename":"Biquipédia"}],"dir":"ltr","localname":"Mirandees"},"198":{"code":"my","name":"မြန်မာဘာသာ","site":[{"url":"https://my.wikipedia.org","dbname":"mywiki","code":"wiki","sitename":"ဝီကီပီးဒီးယား"},{"url":"https://my.wiktionary.org","dbname":"mywiktionary","code":"wiktionary","sitename":"ဝစ်ရှင်နရီ"},{"url":"https://my.wikibooks.org","dbname":"mywikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Birmaans"},"199":{"code":"myv","name":"эрзянь","site":[{"url":"https://myv.wikipedia.org","dbname":"myvwiki","code":"wiki","sitename":"Википедиясь"}],"dir":"ltr","localname":"Erzya"},"200":{"code":"mzn","name":"مازِرونی","site":[{"url":"https://mzn.wikipedia.org","dbname":"mznwiki","code":"wiki","sitename":"ویکی‌پدیا"}],"dir":"rtl","localname":"Masanderani"},"201":{"code":"na","name":"Dorerin Naoero","site":[{"url":"https://na.wikipedia.org","dbname":"nawiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://na.wiktionary.org","dbname":"nawiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://na.wikibooks.org","dbname":"nawikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://na.wikiquote.org","dbname":"nawikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Nauru"},"202":{"code":"nah","name":"Nāhuatl","site":[{"url":"https://nah.wikipedia.org","dbname":"nahwiki","code":"wiki","sitename":"Huiquipedia"},{"url":"https://nah.wiktionary.org","dbname":"nahwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://nah.wikibooks.org","dbname":"nahwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Nāhuatl"},"203":{"code":"nap","name":"Napulitano","site":[{"url":"https://nap.wikipedia.org","dbname":"napwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://nap.wikisource.org","dbname":"napwikisource","code":"wikisource","sitename":"Wikisource"}],"dir":"ltr","localname":"Neapolitaans"},"204":{"code":"nds","name":"Plattdüütsch","site":[{"url":"https://nds.wikipedia.org","dbname":"ndswiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://nds.wiktionary.org","dbname":"ndswiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://nds.wikibooks.org","dbname":"ndswikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://nds.wikiquote.org","dbname":"ndswikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Nederduits"},"205":{"code":"nds-nl","name":"Nedersaksies","site":[{"url":"https://nds-nl.wikipedia.org","dbname":"nds_nlwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Nedersaksies"},"206":{"code":"ne","name":"नेपाली","site":[{"url":"https://ne.wikipedia.org","dbname":"newiki","code":"wiki","sitename":"विकिपिडिया"},{"url":"https://ne.wiktionary.org","dbname":"newiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://ne.wikibooks.org","dbname":"newikibooks","code":"wikibooks","sitename":"विकिपुस्तक"}],"dir":"ltr","localname":"Nepalees"},"207":{"code":"new","name":"नेपाल भाषा","site":[{"url":"https://new.wikipedia.org","dbname":"newwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Newari"},"208":{"code":"ng","name":"Oshiwambo","site":[{"url":"https://ng.wikipedia.org","dbname":"ngwiki","code":"wiki","sitename":"Wikipedia","closed":true}],"dir":"ltr","localname":"Ndonga"},"209":{"code":"nia","name":"Li Niha","site":[{"url":"https://nia.wikipedia.org","dbname":"niawiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://nia.wiktionary.org","dbname":"niawiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Nias"},"210":{"code":"nl","name":"Nederlands","site":[{"url":"https://nl.wikipedia.org","dbname":"nlwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://nl.wiktionary.org","dbname":"nlwiktionary","code":"wiktionary","sitename":"WikiWoordenboek"},{"url":"https://nl.wikibooks.org","dbname":"nlwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://nl.wikinews.org","dbname":"nlwikinews","code":"wikinews","sitename":"Wikinieuws"},{"url":"https://nl.wikiquote.org","dbname":"nlwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://nl.wikisource.org","dbname":"nlwikisource","code":"wikisource","sitename":"Wikisource"},{"url":"https://nl.wikivoyage.org","dbname":"nlwikivoyage","code":"wikivoyage","sitename":"Wikivoyage"}],"dir":"ltr","localname":"Nederlands"},"211":{"code":"nn","name":"norsk nynorsk","site":[{"url":"https://nn.wikipedia.org","dbname":"nnwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://nn.wiktionary.org","dbname":"nnwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://nn.wikiquote.org","dbname":"nnwikiquote","code":"wikiquote","sitename":"Wikiquote"}],"dir":"ltr","localname":"Nuwe Noors"},"212":{"code":"no","name":"norsk","site":[{"url":"https://no.wikipedia.org","dbname":"nowiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://no.wiktionary.org","dbname":"nowiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://no.wikibooks.org","dbname":"nowikibooks","code":"wikibooks","sitename":"Wikibøker"},{"url":"https://no.wikinews.org","dbname":"nowikinews","code":"wikinews","sitename":"Wikinytt"},{"url":"https://no.wikiquote.org","dbname":"nowikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://no.wikisource.org","dbname":"nowikisource","code":"wikisource","sitename":"Wikikilden"}],"dir":"ltr","localname":"Noors"},"213":{"code":"nov","name":"Novial","site":[{"url":"https://nov.wikipedia.org","dbname":"novwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Novial"},"214":{"code":"nqo","name":"ߒߞߏ","site":[{"url":"https://nqo.wikipedia.org","dbname":"nqowiki","code":"wiki","sitename":"ߥߞߌߔߘߋߞߎ"}],"dir":"rtl","localname":"N’Ko"},"215":{"code":"nrm","name":"Nouormand","site":[{"url":"https://nrm.wikipedia.org","dbname":"nrmwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Norman"},"216":{"code":"nso","name":"Sesotho sa Leboa","site":[{"url":"https://nso.wikipedia.org","dbname":"nsowiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Noord-Sotho"},"217":{"code":"nv","name":"Diné bizaad","site":[{"url":"https://nv.wikipedia.org","dbname":"nvwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Navajo"},"218":{"code":"ny","name":"Chi-Chewa","site":[{"url":"https://ny.wikipedia.org","dbname":"nywiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Nyanja"},"219":{"code":"oc","name":"occitan","site":[{"url":"https://oc.wikipedia.org","dbname":"ocwiki","code":"wiki","sitename":"Wikipèdia"},{"url":"https://oc.wiktionary.org","dbname":"ocwiktionary","code":"wiktionary","sitename":"Wikiccionari"},{"url":"https://oc.wikibooks.org","dbname":"ocwikibooks","code":"wikibooks","sitename":"Wikilibres"}],"dir":"ltr","localname":"Oksitaans"},"220":{"code":"olo","name":"livvinkarjala","site":[{"url":"https://olo.wikipedia.org","dbname":"olowiki","code":"wiki","sitename":"Wikipedii"}],"dir":"ltr","localname":"Livvi-Karelian"},"221":{"code":"om","name":"Oromoo","site":[{"url":"https://om.wikipedia.org","dbname":"omwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://om.wiktionary.org","dbname":"omwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Oromo"},"222":{"code":"or","name":"ଓଡ଼ିଆ","site":[{"url":"https://or.wikipedia.org","dbname":"orwiki","code":"wiki","sitename":"ଉଇକିପିଡ଼ିଆ"},{"url":"https://or.wiktionary.org","dbname":"orwiktionary","code":"wiktionary","sitename":"ଉଇକିଅଭିଧାନ"},{"url":"https://or.wikisource.org","dbname":"orwikisource","code":"wikisource","sitename":"ଉଇକିପାଠାଗାର"}],"dir":"ltr","localname":"Oriya"},"223":{"code":"os","name":"ирон","site":[{"url":"https://os.wikipedia.org","dbname":"oswiki","code":"wiki","sitename":"Википеди"}],"dir":"ltr","localname":"Osseties"},"224":{"code":"pa","name":"ਪੰਜਾਬੀ","site":[{"url":"https://pa.wikipedia.org","dbname":"pawiki","code":"wiki","sitename":"ਵਿਕੀਪੀਡੀਆ"},{"url":"https://pa.wiktionary.org","dbname":"pawiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://pa.wikibooks.org","dbname":"pawikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://pa.wikisource.org","dbname":"pawikisource","code":"wikisource","sitename":"ਵਿਕੀਸਰੋਤ"}],"dir":"ltr","localname":"Pandjabi"},"225":{"code":"pag","name":"Pangasinan","site":[{"url":"https://pag.wikipedia.org","dbname":"pagwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Pangasinan"},"226":{"code":"pam","name":"Kapampangan","site":[{"url":"https://pam.wikipedia.org","dbname":"pamwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Pampanga"},"227":{"code":"pap","name":"Papiamentu","site":[{"url":"https://pap.wikipedia.org","dbname":"papwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Papiamento"},"228":{"code":"pcd","name":"Picard","site":[{"url":"https://pcd.wikipedia.org","dbname":"pcdwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Picard"},"229":{"code":"pcm","name":"Naijá","site":[{"url":"https://pcm.wikipedia.org","dbname":"pcmwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Nigeriese Pidgin"},"230":{"code":"pdc","name":"Deitsch","site":[{"url":"https://pdc.wikipedia.org","dbname":"pdcwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Pennsylvania German"},"231":{"code":"pfl","name":"Pälzisch","site":[{"url":"https://pfl.wikipedia.org","dbname":"pflwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Palatine German"},"232":{"code":"pi","name":"पालि","site":[{"url":"https://pi.wikipedia.org","dbname":"piwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://pi.wiktionary.org","dbname":"piwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true}],"dir":"ltr","localname":"Pali"},"233":{"code":"pih","name":"Norfuk / Pitkern","site":[{"url":"https://pih.wikipedia.org","dbname":"pihwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Norfuk / Pitkern"},"234":{"code":"pl","name":"polski","site":[{"url":"https://pl.wikipedia.org","dbname":"plwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://pl.wiktionary.org","dbname":"plwiktionary","code":"wiktionary","sitename":"Wikisłownik"},{"url":"https://pl.wikibooks.org","dbname":"plwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://pl.wikinews.org","dbname":"plwikinews","code":"wikinews","sitename":"Wikinews"},{"url":"https://pl.wikiquote.org","dbname":"plwikiquote","code":"wikiquote","sitename":"Wikicytaty"},{"url":"https://pl.wikisource.org","dbname":"plwikisource","code":"wikisource","sitename":"Wikiźródła"},{"url":"https://pl.wikivoyage.org","dbname":"plwikivoyage","code":"wikivoyage","sitename":"Wikipodróże"}],"dir":"ltr","localname":"Pools"},"235":{"code":"pms","name":"Piemontèis","site":[{"url":"https://pms.wikipedia.org","dbname":"pmswiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://pms.wikisource.org","dbname":"pmswikisource","code":"wikisource","sitename":"Wikisource"}],"dir":"ltr","localname":"Piedmontese"},"236":{"code":"pnb","name":"پنجابی","site":[{"url":"https://pnb.wikipedia.org","dbname":"pnbwiki","code":"wiki","sitename":"وکیپیڈیا"},{"url":"https://pnb.wiktionary.org","dbname":"pnbwiktionary","code":"wiktionary","sitename":"وکشنری"}],"dir":"rtl","localname":"Western Punjabi"},"237":{"code":"pnt","name":"Ποντιακά","site":[{"url":"https://pnt.wikipedia.org","dbname":"pntwiki","code":"wiki","sitename":"Βικιπαίδεια"}],"dir":"ltr","localname":"Pontic"},"238":{"code":"ps","name":"پښتو","site":[{"url":"https://ps.wikipedia.org","dbname":"pswiki","code":"wiki","sitename":"ويکيپېډيا"},{"url":"https://ps.wiktionary.org","dbname":"pswiktionary","code":"wiktionary","sitename":"ويکيسيند"},{"url":"https://ps.wikibooks.org","dbname":"pswikibooks","code":"wikibooks","sitename":"ويکيتابونه","closed":true},{"url":"https://ps.wikivoyage.org","dbname":"pswikivoyage","code":"wikivoyage","sitename":"ويکيسفر"}],"dir":"rtl","localname":"Pasjto"},"239":{"code":"pt","name":"português","site":[{"url":"https://pt.wikipedia.org","dbname":"ptwiki","code":"wiki","sitename":"Wikipédia"},{"url":"https://pt.wiktionary.org","dbname":"ptwiktionary","code":"wiktionary","sitename":"Wikcionário"},{"url":"https://pt.wikibooks.org","dbname":"ptwikibooks","code":"wikibooks","sitename":"Wikilivros"},{"url":"https://pt.wikinews.org","dbname":"ptwikinews","code":"wikinews","sitename":"Wikinotícias"},{"url":"https://pt.wikiquote.org","dbname":"ptwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://pt.wikisource.org","dbname":"ptwikisource","code":"wikisource","sitename":"Wikisource"},{"url":"https://pt.wikiversity.org","dbname":"ptwikiversity","code":"wikiversity","sitename":"Wikiversidade"},{"url":"https://pt.wikivoyage.org","dbname":"ptwikivoyage","code":"wikivoyage","sitename":"Wikivoyage"}],"dir":"ltr","localname":"Portugees"},"240":{"code":"pwn","name":"pinayuanan","site":[{"url":"https://pwn.wikipedia.org","dbname":"pwnwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Paiwan"},"241":{"code":"qu","name":"Runa Simi","site":[{"url":"https://qu.wikipedia.org","dbname":"quwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://qu.wiktionary.org","dbname":"quwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://qu.wikibooks.org","dbname":"quwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://qu.wikiquote.org","dbname":"quwikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Quechua"},"242":{"code":"rm","name":"rumantsch","site":[{"url":"https://rm.wikipedia.org","dbname":"rmwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://rm.wiktionary.org","dbname":"rmwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true},{"url":"https://rm.wikibooks.org","dbname":"rmwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Reto-Romaans"},"243":{"code":"rmy","name":"romani čhib","site":[{"url":"https://rmy.wikipedia.org","dbname":"rmywiki","code":"wiki","sitename":"Vikipidiya"}],"dir":"ltr","localname":"Vlax Romani"},"244":{"code":"rn","name":"ikirundi","site":[{"url":"https://rn.wikipedia.org","dbname":"rnwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://rn.wiktionary.org","dbname":"rnwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true}],"dir":"ltr","localname":"Rundi"},"245":{"code":"ro","name":"română","site":[{"url":"https://ro.wikipedia.org","dbname":"rowiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ro.wiktionary.org","dbname":"rowiktionary","code":"wiktionary","sitename":"Wikționar"},{"url":"https://ro.wikibooks.org","dbname":"rowikibooks","code":"wikibooks","sitename":"Wikimanuale"},{"url":"https://ro.wikinews.org","dbname":"rowikinews","code":"wikinews","sitename":"Wikiștiri"},{"url":"https://ro.wikiquote.org","dbname":"rowikiquote","code":"wikiquote","sitename":"Wikicitat"},{"url":"https://ro.wikisource.org","dbname":"rowikisource","code":"wikisource","sitename":"Wikisource"},{"url":"https://ro.wikivoyage.org","dbname":"rowikivoyage","code":"wikivoyage","sitename":"Wikivoyage"}],"dir":"ltr","localname":"Roemeens"},"246":{"code":"roa-rup","name":"armãneashti","site":[{"url":"https://roa-rup.wikipedia.org","dbname":"roa_rupwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://roa-rup.wiktionary.org","dbname":"roa_rupwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Aromanian"},"247":{"code":"roa-tara","name":"tarandíne","site":[{"url":"https://roa-tara.wikipedia.org","dbname":"roa_tarawiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Tarantino"},"248":{"code":"ru","name":"русский","site":[{"url":"https://ru.wikipedia.org","dbname":"ruwiki","code":"wiki","sitename":"Википедия"},{"url":"https://ru.wiktionary.org","dbname":"ruwiktionary","code":"wiktionary","sitename":"Викисловарь"},{"url":"https://ru.wikibooks.org","dbname":"ruwikibooks","code":"wikibooks","sitename":"Викиучебник"},{"url":"https://ru.wikinews.org","dbname":"ruwikinews","code":"wikinews","sitename":"Викиновости"},{"url":"https://ru.wikiquote.org","dbname":"ruwikiquote","code":"wikiquote","sitename":"Викицитатник"},{"url":"https://ru.wikisource.org","dbname":"ruwikisource","code":"wikisource","sitename":"Викитека"},{"url":"https://ru.wikiversity.org","dbname":"ruwikiversity","code":"wikiversity","sitename":"Викиверситет"},{"url":"https://ru.wikivoyage.org","dbname":"ruwikivoyage","code":"wikivoyage","sitename":"Wikivoyage"}],"dir":"ltr","localname":"Russies"},"249":{"code":"rue","name":"русиньскый","site":[{"url":"https://rue.wikipedia.org","dbname":"ruewiki","code":"wiki","sitename":"Вікіпедія"}],"dir":"ltr","localname":"Rusyn"},"250":{"code":"rw","name":"Ikinyarwanda","site":[{"url":"https://rw.wikipedia.org","dbname":"rwwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://rw.wiktionary.org","dbname":"rwwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Rwandees"},"251":{"code":"sa","name":"संस्कृतम्","site":[{"url":"https://sa.wikipedia.org","dbname":"sawiki","code":"wiki","sitename":"विकिपीडिया"},{"url":"https://sa.wiktionary.org","dbname":"sawiktionary","code":"wiktionary","sitename":"विकिशब्दकोशः"},{"url":"https://sa.wikibooks.org","dbname":"sawikibooks","code":"wikibooks","sitename":"विकिपुस्तकानि"},{"url":"https://sa.wikiquote.org","dbname":"sawikiquote","code":"wikiquote","sitename":"विकिसूक्तिः"},{"url":"https://sa.wikisource.org","dbname":"sawikisource","code":"wikisource","sitename":"विकिस्रोतः"}],"dir":"ltr","localname":"Sanskrit"},"252":{"code":"sah","name":"саха тыла","site":[{"url":"https://sah.wikipedia.org","dbname":"sahwiki","code":"wiki","sitename":"Бикипиэдьийэ"},{"url":"https://sah.wikiquote.org","dbname":"sahwikiquote","code":"wikiquote","sitename":"Биики_Домох"},{"url":"https://sah.wikisource.org","dbname":"sahwikisource","code":"wikisource","sitename":"Бикитиэкэ"}],"dir":"ltr","localname":"Sakhaans"},"253":{"code":"sat","name":"ᱥᱟᱱᱛᱟᱲᱤ","site":[{"url":"https://sat.wikipedia.org","dbname":"satwiki","code":"wiki","sitename":"ᱣᱤᱠᱤᱯᱤᱰᱤᱭᱟ"}],"dir":"ltr","localname":"Santalies"},"254":{"code":"sc","name":"sardu","site":[{"url":"https://sc.wikipedia.org","dbname":"scwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://sc.wiktionary.org","dbname":"scwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true}],"dir":"ltr","localname":"Sardinies"},"255":{"code":"scn","name":"sicilianu","site":[{"url":"https://scn.wikipedia.org","dbname":"scnwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://scn.wiktionary.org","dbname":"scnwiktionary","code":"wiktionary","sitename":"Wikizziunariu"}],"dir":"ltr","localname":"Sisiliaans"},"256":{"code":"sco","name":"Scots","site":[{"url":"https://sco.wikipedia.org","dbname":"scowiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Skots"},"257":{"code":"sd","name":"سنڌي","site":[{"url":"https://sd.wikipedia.org","dbname":"sdwiki","code":"wiki","sitename":"وڪيپيڊيا"},{"url":"https://sd.wiktionary.org","dbname":"sdwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://sd.wikinews.org","dbname":"sdwikinews","code":"wikinews","sitename":"Wikinews","closed":true}],"dir":"rtl","localname":"Sindhi"},"258":{"code":"se","name":"davvisámegiella","site":[{"url":"https://se.wikipedia.org","dbname":"sewiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://se.wikibooks.org","dbname":"sewikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Noord-Sami"},"259":{"code":"sg","name":"Sängö","site":[{"url":"https://sg.wikipedia.org","dbname":"sgwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://sg.wiktionary.org","dbname":"sgwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Sango"},"260":{"code":"sh","name":"srpskohrvatski / српскохрватски","site":[{"url":"https://sh.wikipedia.org","dbname":"shwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://sh.wiktionary.org","dbname":"shwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Serwo-Kroaties"},"261":{"code":"shi","name":"Taclḥit","site":[{"url":"https://shi.wikipedia.org","dbname":"shiwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Tachelhit"},"262":{"code":"shn","name":"ၽႃႇသႃႇတႆး ","site":[{"url":"https://shn.wikipedia.org","dbname":"shnwiki","code":"wiki","sitename":"ဝီႇၶီႇၽီးတီးယႃး"},{"url":"https://shn.wiktionary.org","dbname":"shnwiktionary","code":"wiktionary","sitename":"ဝိၵ်ႇသျိၼ်ႇၼရီႇ"},{"url":"https://shn.wikibooks.org","dbname":"shnwikibooks","code":"wikibooks","sitename":"ဝီႇၶီႇပပ်ႉ"},{"url":"https://shn.wikivoyage.org","dbname":"shnwikivoyage","code":"wikivoyage","sitename":"ဝီႇၶီႇဝွႆးဢဵတ်ႇꩡ်"}],"dir":"ltr","localname":"Shan"},"263":{"code":"shy","name":"tacawit","site":[{"url":"https://shy.wiktionary.org","dbname":"shywiktionary","code":"wiktionary","sitename":"Wikasegzawal"}],"dir":"ltr","localname":"Shawiya"},"264":{"code":"si","name":"සිංහල","site":[{"url":"https://si.wikipedia.org","dbname":"siwiki","code":"wiki","sitename":"විකිපීඩියා"},{"url":"https://si.wiktionary.org","dbname":"siwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://si.wikibooks.org","dbname":"siwikibooks","code":"wikibooks","sitename":"Wikibooks"}],"dir":"ltr","localname":"Sinhala"},"265":{"code":"simple","name":"Simple English","site":[{"url":"https://simple.wikipedia.org","dbname":"simplewiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://simple.wiktionary.org","dbname":"simplewiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://simple.wikibooks.org","dbname":"simplewikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://simple.wikiquote.org","dbname":"simplewikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Simple English"},"266":{"code":"sk","name":"slovenčina","site":[{"url":"https://sk.wikipedia.org","dbname":"skwiki","code":"wiki","sitename":"Wikipédia"},{"url":"https://sk.wiktionary.org","dbname":"skwiktionary","code":"wiktionary","sitename":"Wikislovník"},{"url":"https://sk.wikibooks.org","dbname":"skwikibooks","code":"wikibooks","sitename":"Wikiknihy"},{"url":"https://sk.wikiquote.org","dbname":"skwikiquote","code":"wikiquote","sitename":"Wikicitáty"},{"url":"https://sk.wikisource.org","dbname":"skwikisource","code":"wikisource","sitename":"Wikizdroje"}],"dir":"ltr","localname":"Slowaaks"},"267":{"code":"skr","name":"سرائیکی","site":[{"url":"https://skr.wikipedia.org","dbname":"skrwiki","code":"wiki","sitename":"وکیپیڈیا"},{"url":"https://skr.wiktionary.org","dbname":"skrwiktionary","code":"wiktionary","sitename":"وکشنری"}],"dir":"rtl","localname":"Saraiki"},"268":{"code":"sl","name":"slovenščina","site":[{"url":"https://sl.wikipedia.org","dbname":"slwiki","code":"wiki","sitename":"Wikipedija"},{"url":"https://sl.wiktionary.org","dbname":"slwiktionary","code":"wiktionary","sitename":"Wikislovar"},{"url":"https://sl.wikibooks.org","dbname":"slwikibooks","code":"wikibooks","sitename":"Wikiknjige"},{"url":"https://sl.wikiquote.org","dbname":"slwikiquote","code":"wikiquote","sitename":"Wikinavedek"},{"url":"https://sl.wikisource.org","dbname":"slwikisource","code":"wikisource","sitename":"Wikivir"},{"url":"https://sl.wikiversity.org","dbname":"slwikiversity","code":"wikiversity","sitename":"Wikiverza"}],"dir":"ltr","localname":"Sloweens"},"269":{"code":"sm","name":"Gagana Samoa","site":[{"url":"https://sm.wikipedia.org","dbname":"smwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://sm.wiktionary.org","dbname":"smwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Samoaans"},"270":{"code":"smn","name":"anarâškielâ","site":[{"url":"https://smn.wikipedia.org","dbname":"smnwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Inari Sami"},"271":{"code":"sn","name":"chiShona","site":[{"url":"https://sn.wikipedia.org","dbname":"snwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://sn.wiktionary.org","dbname":"snwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true}],"dir":"ltr","localname":"Shona"},"272":{"code":"so","name":"Soomaaliga","site":[{"url":"https://so.wikipedia.org","dbname":"sowiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://so.wiktionary.org","dbname":"sowiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Somalies"},"273":{"code":"sq","name":"shqip","site":[{"url":"https://sq.wikipedia.org","dbname":"sqwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://sq.wiktionary.org","dbname":"sqwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://sq.wikibooks.org","dbname":"sqwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://sq.wikinews.org","dbname":"sqwikinews","code":"wikinews","sitename":"Wikilajme"},{"url":"https://sq.wikiquote.org","dbname":"sqwikiquote","code":"wikiquote","sitename":"Wikiquote"}],"dir":"ltr","localname":"Albanees"},"274":{"code":"sr","name":"српски / srpski","site":[{"url":"https://sr.wikipedia.org","dbname":"srwiki","code":"wiki","sitename":"Википедија"},{"url":"https://sr.wiktionary.org","dbname":"srwiktionary","code":"wiktionary","sitename":"Викиречник"},{"url":"https://sr.wikibooks.org","dbname":"srwikibooks","code":"wikibooks","sitename":"Викикњиге"},{"url":"https://sr.wikinews.org","dbname":"srwikinews","code":"wikinews","sitename":"Викиновости"},{"url":"https://sr.wikiquote.org","dbname":"srwikiquote","code":"wikiquote","sitename":"Викицитат"},{"url":"https://sr.wikisource.org","dbname":"srwikisource","code":"wikisource","sitename":"Викизворник"}],"dir":"ltr","localname":"Serwies"},"275":{"code":"srn","name":"Sranantongo","site":[{"url":"https://srn.wikipedia.org","dbname":"srnwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Sranan Tongo"},"276":{"code":"ss","name":"SiSwati","site":[{"url":"https://ss.wikipedia.org","dbname":"sswiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ss.wiktionary.org","dbname":"sswiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Swazi"},"277":{"code":"st","name":"Sesotho","site":[{"url":"https://st.wikipedia.org","dbname":"stwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://st.wiktionary.org","dbname":"stwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Suid-Sotho"},"278":{"code":"stq","name":"Seeltersk","site":[{"url":"https://stq.wikipedia.org","dbname":"stqwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Saterland Frisian"},"279":{"code":"su","name":"Sunda","site":[{"url":"https://su.wikipedia.org","dbname":"suwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://su.wiktionary.org","dbname":"suwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://su.wikibooks.org","dbname":"suwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://su.wikiquote.org","dbname":"suwikiquote","code":"wikiquote","sitename":"Wikiquote"}],"dir":"ltr","localname":"Sundanees"},"280":{"code":"sv","name":"svenska","site":[{"url":"https://sv.wikipedia.org","dbname":"svwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://sv.wiktionary.org","dbname":"svwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://sv.wikibooks.org","dbname":"svwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://sv.wikinews.org","dbname":"svwikinews","code":"wikinews","sitename":"Wikinews"},{"url":"https://sv.wikiquote.org","dbname":"svwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://sv.wikisource.org","dbname":"svwikisource","code":"wikisource","sitename":"Wikisource"},{"url":"https://sv.wikiversity.org","dbname":"svwikiversity","code":"wikiversity","sitename":"Wikiversity"},{"url":"https://sv.wikivoyage.org","dbname":"svwikivoyage","code":"wikivoyage","sitename":"Wikivoyage"}],"dir":"ltr","localname":"Sweeds"},"281":{"code":"sw","name":"Kiswahili","site":[{"url":"https://sw.wikipedia.org","dbname":"swwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://sw.wiktionary.org","dbname":"swwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://sw.wikibooks.org","dbname":"swwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Swahili"},"282":{"code":"szl","name":"ślůnski","site":[{"url":"https://szl.wikipedia.org","dbname":"szlwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Silesian"},"283":{"code":"szy","name":"Sakizaya","site":[{"url":"https://szy.wikipedia.org","dbname":"szywiki","code":"wiki","sitename":"Wikipitiya"}],"dir":"ltr","localname":"Sakizaya"},"284":{"code":"ta","name":"தமிழ்","site":[{"url":"https://ta.wikipedia.org","dbname":"tawiki","code":"wiki","sitename":"விக்கிப்பீடியா"},{"url":"https://ta.wiktionary.org","dbname":"tawiktionary","code":"wiktionary","sitename":"விக்சனரி"},{"url":"https://ta.wikibooks.org","dbname":"tawikibooks","code":"wikibooks","sitename":"விக்கிநூல்கள்"},{"url":"https://ta.wikinews.org","dbname":"tawikinews","code":"wikinews","sitename":"விக்கிசெய்தி"},{"url":"https://ta.wikiquote.org","dbname":"tawikiquote","code":"wikiquote","sitename":"விக்கிமேற்கோள்"},{"url":"https://ta.wikisource.org","dbname":"tawikisource","code":"wikisource","sitename":"விக்கிமூலம்"}],"dir":"ltr","localname":"Tamil"},"285":{"code":"tay","name":"Tayal","site":[{"url":"https://tay.wikipedia.org","dbname":"taywiki","code":"wiki","sitename":"Wikipidia"}],"dir":"ltr","localname":"Tayal"},"286":{"code":"tcy","name":"ತುಳು","site":[{"url":"https://tcy.wikipedia.org","dbname":"tcywiki","code":"wiki","sitename":"ವಿಕಿಪೀಡಿಯ"}],"dir":"ltr","localname":"Tulu"},"287":{"code":"te","name":"తెలుగు","site":[{"url":"https://te.wikipedia.org","dbname":"tewiki","code":"wiki","sitename":"వికీపీడియా"},{"url":"https://te.wiktionary.org","dbname":"tewiktionary","code":"wiktionary","sitename":"విక్షనరీ"},{"url":"https://te.wikibooks.org","dbname":"tewikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://te.wikiquote.org","dbname":"tewikiquote","code":"wikiquote","sitename":"వికీవ్యాఖ్య"},{"url":"https://te.wikisource.org","dbname":"tewikisource","code":"wikisource","sitename":"వికీసోర్స్"}],"dir":"ltr","localname":"Teloegoe"},"288":{"code":"tet","name":"tetun","site":[{"url":"https://tet.wikipedia.org","dbname":"tetwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Tetoem"},"289":{"code":"tg","name":"тоҷикӣ","site":[{"url":"https://tg.wikipedia.org","dbname":"tgwiki","code":"wiki","sitename":"Википедиа"},{"url":"https://tg.wiktionary.org","dbname":"tgwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://tg.wikibooks.org","dbname":"tgwikibooks","code":"wikibooks","sitename":"Wikibooks"}],"dir":"ltr","localname":"Tadjiks"},"290":{"code":"th","name":"ไทย","site":[{"url":"https://th.wikipedia.org","dbname":"thwiki","code":"wiki","sitename":"วิกิพีเดีย"},{"url":"https://th.wiktionary.org","dbname":"thwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://th.wikibooks.org","dbname":"thwikibooks","code":"wikibooks","sitename":"วิกิตำรา"},{"url":"https://th.wikinews.org","dbname":"thwikinews","code":"wikinews","sitename":"Wikinews","closed":true},{"url":"https://th.wikiquote.org","dbname":"thwikiquote","code":"wikiquote","sitename":"วิกิคำคม"},{"url":"https://th.wikisource.org","dbname":"thwikisource","code":"wikisource","sitename":"วิกิซอร์ซ"}],"dir":"ltr","localname":"Thai"},"291":{"code":"ti","name":"ትግርኛ","site":[{"url":"https://ti.wikipedia.org","dbname":"tiwiki","code":"wiki","sitename":"ዊኪፔዲያ"},{"url":"https://ti.wiktionary.org","dbname":"tiwiktionary","code":"wiktionary","sitename":"ዊኪ-መዝገበ-ቃላት"}],"dir":"ltr","localname":"Tigrinya"},"292":{"code":"tk","name":"Türkmençe","site":[{"url":"https://tk.wikipedia.org","dbname":"tkwiki","code":"wiki","sitename":"Wikipediýa"},{"url":"https://tk.wiktionary.org","dbname":"tkwiktionary","code":"wiktionary","sitename":"Wikisözlük"},{"url":"https://tk.wikibooks.org","dbname":"tkwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://tk.wikiquote.org","dbname":"tkwikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Turkmeens"},"293":{"code":"tl","name":"Tagalog","site":[{"url":"https://tl.wikipedia.org","dbname":"tlwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://tl.wiktionary.org","dbname":"tlwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://tl.wikibooks.org","dbname":"tlwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://tl.wikiquote.org","dbname":"tlwikiquote","code":"wikiquote","sitename":"Wikiquote"}],"dir":"ltr","localname":"Tagalog"},"294":{"code":"tn","name":"Setswana","site":[{"url":"https://tn.wikipedia.org","dbname":"tnwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://tn.wiktionary.org","dbname":"tnwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Tswana"},"295":{"code":"to","name":"lea faka-Tonga","site":[{"url":"https://to.wikipedia.org","dbname":"towiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://to.wiktionary.org","dbname":"towiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true}],"dir":"ltr","localname":"Tongaans"},"296":{"code":"tpi","name":"Tok Pisin","site":[{"url":"https://tpi.wikipedia.org","dbname":"tpiwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://tpi.wiktionary.org","dbname":"tpiwiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Tok Pisin"},"297":{"code":"tr","name":"Türkçe","site":[{"url":"https://tr.wikipedia.org","dbname":"trwiki","code":"wiki","sitename":"Vikipedi"},{"url":"https://tr.wiktionary.org","dbname":"trwiktionary","code":"wiktionary","sitename":"Vikisözlük"},{"url":"https://tr.wikibooks.org","dbname":"trwikibooks","code":"wikibooks","sitename":"Vikikitap"},{"url":"https://tr.wikinews.org","dbname":"trwikinews","code":"wikinews","sitename":"Vikihaber","closed":true},{"url":"https://tr.wikiquote.org","dbname":"trwikiquote","code":"wikiquote","sitename":"Vikisöz"},{"url":"https://tr.wikisource.org","dbname":"trwikisource","code":"wikisource","sitename":"Vikikaynak"},{"url":"https://tr.wikivoyage.org","dbname":"trwikivoyage","code":"wikivoyage","sitename":"Vikigezgin"}],"dir":"ltr","localname":"Turks"},"298":{"code":"trv","name":"Seediq","site":[{"url":"https://trv.wikipedia.org","dbname":"trvwiki","code":"wiki","sitename":"Wikipidiya"}],"dir":"ltr","localname":"Taroko"},"299":{"code":"ts","name":"Xitsonga","site":[{"url":"https://ts.wikipedia.org","dbname":"tswiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ts.wiktionary.org","dbname":"tswiktionary","code":"wiktionary","sitename":"Wiktionary"}],"dir":"ltr","localname":"Tsonga"},"300":{"code":"tt","name":"татарча/tatarça","site":[{"url":"https://tt.wikipedia.org","dbname":"ttwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://tt.wiktionary.org","dbname":"ttwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://tt.wikibooks.org","dbname":"ttwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://tt.wikiquote.org","dbname":"ttwikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Tataars"},"301":{"code":"tum","name":"chiTumbuka","site":[{"url":"https://tum.wikipedia.org","dbname":"tumwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Toemboeka"},"302":{"code":"tw","name":"Twi","site":[{"url":"https://tw.wikipedia.org","dbname":"twwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://tw.wiktionary.org","dbname":"twwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true}],"dir":"ltr","localname":"Twi"},"303":{"code":"ty","name":"reo tahiti","site":[{"url":"https://ty.wikipedia.org","dbname":"tywiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Tahities"},"304":{"code":"tyv","name":"тыва дыл","site":[{"url":"https://tyv.wikipedia.org","dbname":"tyvwiki","code":"wiki","sitename":"Википедия"}],"dir":"ltr","localname":"Tuvinees"},"305":{"code":"udm","name":"удмурт","site":[{"url":"https://udm.wikipedia.org","dbname":"udmwiki","code":"wiki","sitename":"Википедия"}],"dir":"ltr","localname":"Udmurt"},"306":{"code":"ug","name":"ئۇيغۇرچە / Uyghurche","site":[{"url":"https://ug.wikipedia.org","dbname":"ugwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://ug.wiktionary.org","dbname":"ugwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://ug.wikibooks.org","dbname":"ugwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://ug.wikiquote.org","dbname":"ugwikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"rtl","localname":"Uighur"},"307":{"code":"uk","name":"українська","site":[{"url":"https://uk.wikipedia.org","dbname":"ukwiki","code":"wiki","sitename":"Вікіпедія"},{"url":"https://uk.wiktionary.org","dbname":"ukwiktionary","code":"wiktionary","sitename":"Вікісловник"},{"url":"https://uk.wikibooks.org","dbname":"ukwikibooks","code":"wikibooks","sitename":"Вікіпідручник"},{"url":"https://uk.wikinews.org","dbname":"ukwikinews","code":"wikinews","sitename":"Вікіновини"},{"url":"https://uk.wikiquote.org","dbname":"ukwikiquote","code":"wikiquote","sitename":"Вікіцитати"},{"url":"https://uk.wikisource.org","dbname":"ukwikisource","code":"wikisource","sitename":"Вікіджерела"},{"url":"https://uk.wikivoyage.org","dbname":"ukwikivoyage","code":"wikivoyage","sitename":"Вікімандри"}],"dir":"ltr","localname":"Oekraïens"},"308":{"code":"ur","name":"اردو","site":[{"url":"https://ur.wikipedia.org","dbname":"urwiki","code":"wiki","sitename":"ویکیپیڈیا"},{"url":"https://ur.wiktionary.org","dbname":"urwiktionary","code":"wiktionary","sitename":"ویکی لغت"},{"url":"https://ur.wikibooks.org","dbname":"urwikibooks","code":"wikibooks","sitename":"ویکی کتب"},{"url":"https://ur.wikiquote.org","dbname":"urwikiquote","code":"wikiquote","sitename":"ویکی اقتباس"}],"dir":"rtl","localname":"Oerdoe"},"309":{"code":"uz","name":"oʻzbekcha/ўзбекча","site":[{"url":"https://uz.wikipedia.org","dbname":"uzwiki","code":"wiki","sitename":"Vikipediya"},{"url":"https://uz.wiktionary.org","dbname":"uzwiktionary","code":"wiktionary","sitename":"Vikilug‘at"},{"url":"https://uz.wikibooks.org","dbname":"uzwikibooks","code":"wikibooks","sitename":"Vikikitob","closed":true},{"url":"https://uz.wikiquote.org","dbname":"uzwikiquote","code":"wikiquote","sitename":"Vikiiqtibos"}],"dir":"ltr","localname":"Oezbeeks"},"310":{"code":"ve","name":"Tshivenda","site":[{"url":"https://ve.wikipedia.org","dbname":"vewiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Venda"},"311":{"code":"vec","name":"vèneto","site":[{"url":"https://vec.wikipedia.org","dbname":"vecwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://vec.wiktionary.org","dbname":"vecwiktionary","code":"wiktionary","sitename":"Wikisionario"},{"url":"https://vec.wikisource.org","dbname":"vecwikisource","code":"wikisource","sitename":"Wikisource"}],"dir":"ltr","localname":"Venetian"},"312":{"code":"vep","name":"vepsän kel’","site":[{"url":"https://vep.wikipedia.org","dbname":"vepwiki","code":"wiki","sitename":"Vikipedii"}],"dir":"ltr","localname":"Veps"},"313":{"code":"vi","name":"Tiếng Việt","site":[{"url":"https://vi.wikipedia.org","dbname":"viwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://vi.wiktionary.org","dbname":"viwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://vi.wikibooks.org","dbname":"viwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://vi.wikiquote.org","dbname":"viwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://vi.wikisource.org","dbname":"viwikisource","code":"wikisource","sitename":"Wikisource"},{"url":"https://vi.wikivoyage.org","dbname":"viwikivoyage","code":"wikivoyage","sitename":"Wikivoyage"}],"dir":"ltr","localname":"Viëtnamees"},"314":{"code":"vls","name":"West-Vlams","site":[{"url":"https://vls.wikipedia.org","dbname":"vlswiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"West Flemish"},"315":{"code":"vo","name":"Volapük","site":[{"url":"https://vo.wikipedia.org","dbname":"vowiki","code":"wiki","sitename":"Vükiped"},{"url":"https://vo.wiktionary.org","dbname":"vowiktionary","code":"wiktionary","sitename":"Vükivödabuk"},{"url":"https://vo.wikibooks.org","dbname":"vowikibooks","code":"wikibooks","sitename":"Vükibuks","closed":true},{"url":"https://vo.wikiquote.org","dbname":"vowikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Volapük"},"316":{"code":"wa","name":"walon","site":[{"url":"https://wa.wikipedia.org","dbname":"wawiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://wa.wiktionary.org","dbname":"wawiktionary","code":"wiktionary","sitename":"Wiccionaire"},{"url":"https://wa.wikibooks.org","dbname":"wawikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://wa.wikisource.org","dbname":"wawikisource","code":"wikisource","sitename":"Wikisource"}],"dir":"ltr","localname":"Walloon"},"317":{"code":"war","name":"Winaray","site":[{"url":"https://war.wikipedia.org","dbname":"warwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Waray"},"318":{"code":"wo","name":"Wolof","site":[{"url":"https://wo.wikipedia.org","dbname":"wowiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://wo.wiktionary.org","dbname":"wowiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://wo.wikiquote.org","dbname":"wowikiquote","code":"wikiquote","sitename":"Wikiquote"}],"dir":"ltr","localname":"Wolof"},"319":{"code":"wuu","name":"吴语","site":[{"url":"https://wuu.wikipedia.org","dbname":"wuuwiki","code":"wiki","sitename":"维基百科"}],"dir":"ltr","localname":"Wu-Sjinees"},"320":{"code":"xal","name":"хальмг","site":[{"url":"https://xal.wikipedia.org","dbname":"xalwiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Kalmyk"},"321":{"code":"xh","name":"isiXhosa","site":[{"url":"https://xh.wikipedia.org","dbname":"xhwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://xh.wiktionary.org","dbname":"xhwiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true},{"url":"https://xh.wikibooks.org","dbname":"xhwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Xhosa"},"322":{"code":"xmf","name":"მარგალური","site":[{"url":"https://xmf.wikipedia.org","dbname":"xmfwiki","code":"wiki","sitename":"ვიკიპედია"}],"dir":"ltr","localname":"Mingrelian"},"323":{"code":"yi","name":"ייִדיש","site":[{"url":"https://yi.wikipedia.org","dbname":"yiwiki","code":"wiki","sitename":"װיקיפּעדיע"},{"url":"https://yi.wiktionary.org","dbname":"yiwiktionary","code":"wiktionary","sitename":"װיקיװערטערבוך"},{"url":"https://yi.wikisource.org","dbname":"yiwikisource","code":"wikisource","sitename":"װיקיביבליאָטעק"}],"dir":"rtl","localname":"Jiddisj"},"324":{"code":"yo","name":"Yorùbá","site":[{"url":"https://yo.wikipedia.org","dbname":"yowiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://yo.wiktionary.org","dbname":"yowiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true},{"url":"https://yo.wikibooks.org","dbname":"yowikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Yoruba"},"325":{"code":"yue","name":"粵語","site":[{"url":"https://yue.wiktionary.org","dbname":"yuewiktionary","code":"wiktionary","sitename":"維基辭典"}],"dir":"ltr","localname":"Kantonees"},"326":{"code":"za","name":"Vahcuengh","site":[{"url":"https://za.wikipedia.org","dbname":"zawiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://za.wiktionary.org","dbname":"zawiktionary","code":"wiktionary","sitename":"Wiktionary","closed":true},{"url":"https://za.wikibooks.org","dbname":"zawikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://za.wikiquote.org","dbname":"zawikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true}],"dir":"ltr","localname":"Zhuang"},"327":{"code":"zea","name":"Zeêuws","site":[{"url":"https://zea.wikipedia.org","dbname":"zeawiki","code":"wiki","sitename":"Wikipedia"}],"dir":"ltr","localname":"Zeelandic"},"328":{"code":"zh","name":"中文","site":[{"url":"https://zh.wikipedia.org","dbname":"zhwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://zh.wiktionary.org","dbname":"zhwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://zh.wikibooks.org","dbname":"zhwikibooks","code":"wikibooks","sitename":"Wikibooks"},{"url":"https://zh.wikinews.org","dbname":"zhwikinews","code":"wikinews","sitename":"Wikinews"},{"url":"https://zh.wikiquote.org","dbname":"zhwikiquote","code":"wikiquote","sitename":"Wikiquote"},{"url":"https://zh.wikisource.org","dbname":"zhwikisource","code":"wikisource","sitename":"Wikisource"},{"url":"https://zh.wikiversity.org","dbname":"zhwikiversity","code":"wikiversity","sitename":"維基學院"},{"url":"https://zh.wikivoyage.org","dbname":"zhwikivoyage","code":"wikivoyage","sitename":"维基导游"}],"dir":"ltr","localname":"Chinees"},"329":{"code":"zh-classical","name":"文言","site":[{"url":"https://zh-classical.wikipedia.org","dbname":"zh_classicalwiki","code":"wiki","sitename":"維基大典"}],"dir":"ltr","localname":"Classical Chinese"},"330":{"code":"zh-min-nan","name":"Bân-lâm-gú","site":[{"url":"https://zh-min-nan.wikipedia.org","dbname":"zh_min_nanwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://zh-min-nan.wiktionary.org","dbname":"zh_min_nanwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://zh-min-nan.wikibooks.org","dbname":"zh_min_nanwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true},{"url":"https://zh-min-nan.wikiquote.org","dbname":"zh_min_nanwikiquote","code":"wikiquote","sitename":"Wikiquote","closed":true},{"url":"https://zh-min-nan.wikisource.org","dbname":"zh_min_nanwikisource","code":"wikisource","sitename":"Wiki Tô·-su-kóan"}],"dir":"ltr","localname":"Chinese (Min Nan)"},"331":{"code":"zh-yue","name":"粵語","site":[{"url":"https://zh-yue.wikipedia.org","dbname":"zh_yuewiki","code":"wiki","sitename":"維基百科"}],"dir":"ltr","localname":"Cantonese"},"332":{"code":"zu","name":"isiZulu","site":[{"url":"https://zu.wikipedia.org","dbname":"zuwiki","code":"wiki","sitename":"Wikipedia"},{"url":"https://zu.wiktionary.org","dbname":"zuwiktionary","code":"wiktionary","sitename":"Wiktionary"},{"url":"https://zu.wikibooks.org","dbname":"zuwikibooks","code":"wikibooks","sitename":"Wikibooks","closed":true}],"dir":"ltr","localname":"Zoeloe"},"specials":[{"url":"https://advisors.wikimedia.org","dbname":"advisorswiki","code":"advisors","lang":"advisors","sitename":"Advisors","private":true},{"url":"https://advisory.wikimedia.org","dbname":"advisorywiki","code":"advisory","lang":"en","sitename":"Advisory Board","closed":true},{"url":"https://am.wikimedia.org","dbname":"amwikimedia","code":"amwikimedia","lang":"hy","sitename":"Վիքիմեդիա Հայաստան","fishbowl":true},{"url":"https://api.wikimedia.org","dbname":"apiportalwiki","code":"apiportal","lang":"en","sitename":"API Portal"},{"url":"https://ar.wikimedia.org","dbname":"arwikimedia","code":"arwikimedia","lang":"es","sitename":"Wikimedia Argentina"},{"url":"https://arbcom-cs.wikipedia.org","dbname":"arbcom_cswiki","code":"arbcom-cs","lang":"cs","sitename":"Arbitrážní výbor","private":true},{"url":"https://arbcom-de.wikipedia.org","dbname":"arbcom_dewiki","code":"arbcom-de","lang":"de","sitename":"Arbitration Committee","private":true},{"url":"https://arbcom-en.wikipedia.org","dbname":"arbcom_enwiki","code":"arbcom-en","lang":"en","sitename":"Arbitration Committee","private":true},{"url":"https://arbcom-fi.wikipedia.org","dbname":"arbcom_fiwiki","code":"arbcom-fi","lang":"fi","sitename":"Arbitration Committee","private":true},{"url":"https://arbcom-nl.wikipedia.org","dbname":"arbcom_nlwiki","code":"arbcom-nl","lang":"nl","sitename":"Arbitration Committee","private":true},{"url":"https://arbcom-ru.wikipedia.org","dbname":"arbcom_ruwiki","code":"arbcom-ru","lang":"ru","sitename":"Арбитражный комитет русской Википедии","private":true},{"url":"https://auditcom.wikimedia.org","dbname":"auditcomwiki","code":"auditcom","lang":"en","sitename":"Audit Committee","private":true},{"url":"https://bd.wikimedia.org","dbname":"bdwikimedia","code":"bdwikimedia","lang":"bn","sitename":"উইকিমিডিয়া বাংলাদেশ"},{"url":"https://be.wikimedia.org","dbname":"bewikimedia","code":"bewikimedia","lang":"en","sitename":"Wikimedia Belgium"},{"url":"https://beta.wikiversity.org","dbname":"betawikiversity","code":"betawikiversity","lang":"en","sitename":"Wikiversity"},{"url":"https://board.wikimedia.org","dbname":"boardwiki","code":"board","lang":"board","sitename":"Board","private":true},{"url":"https://boardgovcom.wikimedia.org","dbname":"boardgovcomwiki","code":"boardgovcom","lang":"boardgovcom","sitename":"Board Governance Committee","private":true},{"url":"https://br.wikimedia.org","dbname":"brwikimedia","code":"brwikimedia","lang":"pt-BR","sitename":"Wikimedia Brasil"},{"url":"https://ca.wikimedia.org","dbname":"cawikimedia","code":"cawikimedia","lang":"en","sitename":"Wikimedia Canada"},{"url":"https://chair.wikimedia.org","dbname":"chairwiki","code":"chair","lang":"en","sitename":"Wikimedia Board Chair","private":true},{"url":"https://affcom.wikimedia.org","dbname":"chapcomwiki","code":"chapcom","lang":"en","sitename":"Affcom","private":true},{"url":"https://checkuser.wikimedia.org","dbname":"checkuserwiki","code":"checkuser","lang":"en","sitename":"CheckUser Wiki","private":true},{"url":"https://cn.wikimedia.org","dbname":"cnwikimedia","code":"cnwikimedia","lang":"zh","sitename":"中国维基媒体用户组","fishbowl":true},{"url":"https://co.wikimedia.org","dbname":"cowikimedia","code":"cowikimedia","lang":"es","sitename":"Wikimedia Colombia"},{"url":"https://collab.wikimedia.org","dbname":"collabwiki","code":"collab","lang":"en","sitename":"Collab","private":true},{"url":"https://commons.wikimedia.org","dbname":"commonswiki","code":"commons","lang":"commons","sitename":"Wikimedia Commons"},{"url":"https://dk.wikimedia.org","dbname":"dkwikimedia","code":"dkwikimedia","lang":"da","sitename":"Wikimedia Danmark"},{"url":"https://donate.wikimedia.org","dbname":"donatewiki","code":"donate","lang":"en","sitename":"Donate","fishbowl":true},{"url":"https://ec.wikimedia.org","dbname":"ecwikimedia","code":"ecwikimedia","lang":"es","sitename":"Wikimedistas de Ecuador","private":true},{"url":"https://electcom.wikimedia.org","dbname":"electcomwiki","code":"electcom","lang":"electcom","sitename":"Wikipedia","private":true},{"url":"https://ee.wikimedia.org","dbname":"etwikimedia","code":"etwikimedia","lang":"et","sitename":"Wikimedia Eesti"},{"url":"https://exec.wikimedia.org","dbname":"execwiki","code":"exec","lang":"en","sitename":"Wikimedia Executive","private":true},{"url":"https://fdc.wikimedia.org","dbname":"fdcwiki","code":"fdc","lang":"en","sitename":"Wikimedia FDC","private":true},{"url":"https://fi.wikimedia.org","dbname":"fiwikimedia","code":"fiwikimedia","lang":"fi","sitename":"Wikimedia Suomi"},{"url":"https://foundation.wikimedia.org","dbname":"foundationwiki","code":"foundation","lang":"foundation","sitename":"Wikimedia Foundation Governance Wiki"},{"url":"https://ge.wikimedia.org","dbname":"gewikimedia","code":"gewikimedia","lang":"ka","sitename":"ვიკიმედიის მომხმარებელთა საქართველოს ჯგუფი","fishbowl":true},{"url":"https://gr.wikimedia.org","dbname":"grwikimedia","code":"grwikimedia","lang":"el","sitename":"Wikimedia User Group Greece","fishbowl":true},{"url":"https://grants.wikimedia.org","dbname":"grantswiki","code":"grants","lang":"en","sitename":"Wikimedia Foundation Grants Discussion","private":true},{"url":"https://hi.wikimedia.org","dbname":"hiwikimedia","code":"hiwikimedia","lang":"hi","sitename":"Hindi Wikimedians User Group","fishbowl":true},{"url":"https://id.wikimedia.org","dbname":"idwikimedia","code":"idwikimedia","lang":"id","sitename":"Wikimedia Indonesia","fishbowl":true},{"url":"https://id-internal.wikimedia.org","dbname":"id_internalwikimedia","code":"id-internalwikimedia","lang":"id","sitename":"Wikimedia Indonesia (internal)","private":true},{"url":"https://iegcom.wikimedia.org","dbname":"iegcomwiki","code":"iegcom","lang":"en","sitename":"Individual Engagement Grants Committee Wiki","private":true},{"url":"https://il.wikimedia.org","dbname":"ilwikimedia","code":"ilwikimedia","lang":"he","sitename":"ויקימדיה","private":true},{"url":"https://incubator.wikimedia.org","dbname":"incubatorwiki","code":"incubator","lang":"en","sitename":"Wikimedia Incubator"},{"url":"https://internal.wikimedia.org","dbname":"internalwiki","code":"internal","lang":"internal","sitename":"Internal","private":true,"closed":true},{"url":"https://wikitech.wikimedia.org","dbname":"labswiki","code":"labs","lang":"labs","sitename":"Wikipedia","nonglobal":true},{"url":"https://labtestwikitech.wikimedia.org","dbname":"labtestwiki","code":"labtest","lang":"labtest","sitename":"Wikipedia","nonglobal":true},{"url":"https://legalteam.wikimedia.org","dbname":"legalteamwiki","code":"legalteam","lang":"en","sitename":"Legal Team Wiki","private":true},{"url":"https://login.wikimedia.org","dbname":"loginwiki","code":"login","lang":"login","sitename":"Wikimedia Login Wiki"},{"url":"https://mai.wikimedia.org","dbname":"maiwikimedia","code":"maiwikimedia","lang":"mai","sitename":"मैथिली विकिमिडियन्स","fishbowl":true},{"url":"https://www.mediawiki.org","dbname":"mediawikiwiki","code":"mediawiki","lang":"mediawiki","sitename":"MediaWiki"},{"url":"https://meta.wikimedia.org","dbname":"metawiki","code":"meta","lang":"meta","sitename":"Meta"},{"url":"https://mk.wikimedia.org","dbname":"mkwikimedia","code":"mkwikimedia","lang":"mk","sitename":"Викимедија Македонија"},{"url":"https://movementroles.wikimedia.org","dbname":"movementroleswiki","code":"movementroles","lang":"movementroles","sitename":"Movement Roles","private":true},{"url":"https://mx.wikimedia.org","dbname":"mxwikimedia","code":"mxwikimedia","lang":"es","sitename":"Wikimedia México"},{"url":"https://ng.wikimedia.org","dbname":"ngwikimedia","code":"ngwikimedia","lang":"en","sitename":"Wikimedia Nigeria","fishbowl":true},{"url":"https://nl.wikimedia.org","dbname":"nlwikimedia","code":"nlwikimedia","lang":"nl","sitename":"Wikimedia"},{"url":"https://no.wikimedia.org","dbname":"nowikimedia","code":"nowikimedia","lang":"nb","sitename":"Wikimedia Norge"},{"url":"https://noboard-chapters.wikimedia.org","dbname":"noboard_chapterswikimedia","code":"noboard-chapterswikimedia","lang":"nb","sitename":"Wikimedia Norway Internal Board","private":true},{"url":"https://nostalgia.wikipedia.org","dbname":"nostalgiawiki","code":"nostalgia","lang":"nostalgia","sitename":"Wikipedia","fishbowl":true},{"url":"https://nyc.wikimedia.org","dbname":"nycwikimedia","code":"nycwikimedia","lang":"en","sitename":"Wikimedia New York City"},{"url":"https://nz.wikimedia.org","dbname":"nzwikimedia","code":"nzwikimedia","lang":"en","sitename":"Wikimedia","closed":true},{"url":"https://office.wikimedia.org","dbname":"officewiki","code":"office","lang":"en","sitename":"Wikimedia Office","private":true},{"url":"https://ombuds.wikimedia.org","dbname":"ombudsmenwiki","code":"ombudsmen","lang":"en","sitename":"Ombuds Wiki","private":true},{"url":"https://vrt-wiki.wikimedia.org","dbname":"otrs_wikiwiki","code":"otrs-wiki","lang":"en","sitename":"VRT Wiki","private":true},{"url":"https://outreach.wikimedia.org","dbname":"outreachwiki","code":"outreach","lang":"outreach","sitename":"Outreach Wiki"},{"url":"https://pa-us.wikimedia.org","dbname":"pa_uswikimedia","code":"pa-uswikimedia","lang":"en","sitename":"Wikimedia Pennsylvania","closed":true},{"url":"https://pl.wikimedia.org","dbname":"plwikimedia","code":"plwikimedia","lang":"pl","sitename":"Wikimedia"},{"url":"https://projectcom.wikimedia.org","dbname":"projectcomwiki","code":"projectcom","lang":"en","sitename":"Project Grants Committee","private":true},{"url":"https://pt.wikimedia.org","dbname":"ptwikimedia","code":"ptwikimedia","lang":"pt","sitename":"Wikimedia Portugal"},{"url":"https://punjabi.wikimedia.org","dbname":"punjabiwikimedia","code":"punjabiwikimedia","lang":"pa","sitename":"Punjabi Wikimedians","fishbowl":true},{"url":"https://quality.wikimedia.org","dbname":"qualitywiki","code":"quality","lang":"en","sitename":"Wikimedia Quality","closed":true},{"url":"https://romd.wikimedia.org","dbname":"romdwikimedia","code":"romdwikimedia","lang":"ro","sitename":"Wikimedia","fishbowl":true},{"url":"https://rs.wikimedia.org","dbname":"rswikimedia","code":"rswikimedia","lang":"sr","sitename":"Викимедија","fishbowl":true},{"url":"https://ru.wikimedia.org","dbname":"ruwikimedia","code":"ruwikimedia","lang":"ru","sitename":"Викимедиа"},{"url":"https://se.wikimedia.org","dbname":"sewikimedia","code":"sewikimedia","lang":"sv","sitename":"Wikimedia"},{"url":"https://searchcom.wikimedia.org","dbname":"searchcomwiki","code":"searchcom","lang":"en","sitename":"Search Committee","private":true,"closed":true},{"url":"https://wikisource.org","dbname":"sourceswiki","code":"sources","lang":"sources","sitename":"Wikisource"},{"url":"https://spcom.wikimedia.org","dbname":"spcomwiki","code":"spcom","lang":"spcom","sitename":"Spcom","private":true,"closed":true},{"url":"https://species.wikimedia.org","dbname":"specieswiki","code":"species","lang":"species","sitename":"Wikispecies"},{"url":"https://steward.wikimedia.org","dbname":"stewardwiki","code":"steward","lang":"en","sitename":"Steward Wiki","private":true},{"url":"https://strategy.wikimedia.org","dbname":"strategywiki","code":"strategy","lang":"en","sitename":"Strategic Planning","closed":true},{"url":"https://sysop-it.wikipedia.org","dbname":"sysop_itwiki","code":"sysop-it","lang":"it","sitename":"Italian Wikipedia sysops wiki","private":true},{"url":"https://techconduct.wikimedia.org","dbname":"techconductwiki","code":"techconduct","lang":"techconduct","sitename":"CoC committee","private":true},{"url":"https://ten.wikipedia.org","dbname":"tenwiki","code":"ten","lang":"en","sitename":"Wikipedia 10","closed":true},{"url":"https://test.wikipedia.org","dbname":"testwiki","code":"test","lang":"en","sitename":"Wikipedia"},{"url":"https://test2.wikipedia.org","dbname":"test2wiki","code":"test2","lang":"en","sitename":"Wikipedia"},{"url":"https://test-commons.wikimedia.org","dbname":"testcommonswiki","code":"testcommons","lang":"testcommons","sitename":"Test Wikimedia Commons"},{"url":"https://test.wikidata.org","dbname":"testwikidatawiki","code":"testwikidata","lang":"testwikidata","sitename":"Wikipedia"},{"url":"https://thankyou.wikipedia.org","dbname":"thankyouwiki","code":"thankyou","lang":"en","sitename":"Thank You","fishbowl":true},{"url":"https://tr.wikimedia.org","dbname":"trwikimedia","code":"trwikimedia","lang":"tr","sitename":"Wikimedia Türkiye"},{"url":"https://transitionteam.wikimedia.org","dbname":"transitionteamwiki","code":"transitionteam","lang":"en","sitename":"Transition Team Wiki","private":true,"closed":true},{"url":"https://ua.wikimedia.org","dbname":"uawikimedia","code":"uawikimedia","lang":"uk","sitename":"Вікімедіа Україна"},{"url":"https://usability.wikimedia.org","dbname":"usabilitywiki","code":"usability","lang":"en","sitename":"Wikimedia Usability Initiative","closed":true},{"url":"https://vote.wikimedia.org","dbname":"votewiki","code":"vote","lang":"en","sitename":"Wikimedia Vote Wiki","fishbowl":true},{"url":"https://wb.wikimedia.org","dbname":"wbwikimedia","code":"wbwikimedia","lang":"bn","sitename":"West Bengal Wikimedians","fishbowl":true},{"url":"https://wg-en.wikipedia.org","dbname":"wg_enwiki","code":"wg-en","lang":"en","sitename":"Wikipedia Working Group","private":true},{"url":"https://www.wikidata.org","dbname":"wikidatawiki","code":"wikidata","lang":"wikidata","sitename":"Wikipedia"},{"url":"https://wikimania.wikimedia.org","dbname":"wikimaniawiki","code":"wikimania","lang":"wikimania","sitename":"Wikipedia"},{"url":"https://wikimania2005.wikimedia.org","dbname":"wikimania2005wiki","code":"wikimania2005","lang":"wikimania2005","sitename":"Wikipedia","closed":true},{"url":"https://wikimania2006.wikimedia.org","dbname":"wikimania2006wiki","code":"wikimania2006","lang":"wikimania2006","sitename":"Wikipedia","closed":true},{"url":"https://wikimania2007.wikimedia.org","dbname":"wikimania2007wiki","code":"wikimania2007","lang":"wikimania2007","sitename":"Wikipedia","closed":true},{"url":"https://wikimania2008.wikimedia.org","dbname":"wikimania2008wiki","code":"wikimania2008","lang":"wikimania2008","sitename":"Wikipedia","closed":true},{"url":"https://wikimania2009.wikimedia.org","dbname":"wikimania2009wiki","code":"wikimania2009","lang":"wikimania2009","sitename":"Wikipedia","closed":true},{"url":"https://wikimania2010.wikimedia.org","dbname":"wikimania2010wiki","code":"wikimania2010","lang":"wikimania2010","sitename":"Wikipedia","closed":true},{"url":"https://wikimania2011.wikimedia.org","dbname":"wikimania2011wiki","code":"wikimania2011","lang":"wikimania2011","sitename":"Wikipedia","closed":true},{"url":"https://wikimania2012.wikimedia.org","dbname":"wikimania2012wiki","code":"wikimania2012","lang":"wikimania2012","sitename":"Wikipedia","closed":true},{"url":"https://wikimania2013.wikimedia.org","dbname":"wikimania2013wiki","code":"wikimania2013","lang":"wikimania2013","sitename":"Wikipedia","closed":true},{"url":"https://wikimania2014.wikimedia.org","dbname":"wikimania2014wiki","code":"wikimania2014","lang":"wikimania2014","sitename":"Wikipedia","closed":true},{"url":"https://wikimania2015.wikimedia.org","dbname":"wikimania2015wiki","code":"wikimania2015","lang":"wikimania2015","sitename":"Wikipedia","closed":true},{"url":"https://wikimania2016.wikimedia.org","dbname":"wikimania2016wiki","code":"wikimania2016","lang":"wikimania2016","sitename":"Wikipedia","closed":true},{"url":"https://wikimania2017.wikimedia.org","dbname":"wikimania2017wiki","code":"wikimania2017","lang":"wikimania2017","sitename":"Wikipedia","closed":true},{"url":"https://wikimania2018.wikimedia.org","dbname":"wikimania2018wiki","code":"wikimania2018","lang":"wikimania2018","sitename":"Wikipedia","closed":true},{"url":"https://wikimaniateam.wikimedia.org","dbname":"wikimaniateamwiki","code":"wikimaniateam","lang":"en","sitename":"WikimaniaTeam","private":true}]}}`
	getUsersResponse     = `{"batchcomplete":true,"query":{"users":[{"userid":2,"name":"AxelBoldt","editcount":43615,"registration":"2001-07-26T14:50:09Z","groups":["sysop","*","user","autoconfirmed"],"groupmemberships":[{"group":"sysop","expiry":"infinity"}],"emailable":true},{"userid":3,"name":"Tobias Hoevekamp","editcount":2321,"registration":"2001-03-26T20:21:05Z","groups":["extendedconfirmed","*","user","autoconfirmed"],"groupmemberships":[{"group":"extendedconfirmed","expiry":"infinity"}],"emailable":false},{"userid":5,"name":"Hoevekam~enwiki","editcount":14,"registration":"2003-07-09T19:21:37Z","groups":["*","user","autoconfirmed"],"groupmemberships":[],"emailable":false}]}}`
)

const liftWingPayload = `
{
    "output": {
        "prediction": false,
        "probabilities": {
            "true": 0.05083514377474785,
            "false": 0.9491648562252522
        }
    }
}
`

func createActionsAPIServer(sts int, pld string) *httptest.Server {
	rtr := http.NewServeMux()

	rtr.HandleFunc("/w/api.php", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(sts)
		_, _ = w.Write([]byte(pld))
	})

	return httptest.NewServer(rtr)
}

type allPagesTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	pgs map[string]*Page
}

func mockTracer(context.Context, map[string]string) (func(error, string), context.Context) {
	return func(error, string) {}, context.Background()
}

func (s *allPagesTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer:			 mockTracer,
	}

	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	s.pgs = map[string]*Page{}

	if rsp.Query != nil {
		for _, pge := range rsp.Query.AllPages {
			s.pgs[pge.Title] = pge
		}
	}
}

func (s *allPagesTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *allPagesTestSuite) TestGetAllPages() {
	cbk := func(pgs []*Page) {
		for _, pge := range pgs {
			s.Assert().Equal(s.pgs[pge.Title], pge)
		}
	}
	err := s.clt.GetAllPages(s.ctx, s.dtb, cbk)

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestAllPages(t *testing.T) {
	for _, testCase := range []*allPagesTestSuite{
		{
			hse: false,
			pld: `{
				"batchcomplete": true,
				"query": {
					"allpages": [
						{ "pageid": 341265, "ns": 0, "title": "Jungle" },
						{ "pageid": 56461312, "ns": 0, "title": "Jungle-runner" },
						{ "pageid": 18698572, "ns": 0, "title": "Jungle/Drum n bass" },
						{ "pageid": 1487899, "ns": 0, "title": "Jungle2jungle" },
						{ "pageid": 30391179, "ns": 0, "title": "JunglePup" },
						{ "pageid": 18470226, "ns": 0, "title": "Jungle (2000 film)" },
						{ "pageid": 52695829, "ns": 0, "title": "Jungle (2017 film)" },
						{ "pageid": 54000049, "ns": 0, "title": "Jungle (A Boogie wit da Hoodie song)" },
						{ "pageid": 39013520, "ns": 0, "title": "Jungle (Andre Nickatina song)" },
						{ "pageid": 43710655, "ns": 0, "title": "Jungle (Bakufu Slump album)" }
					]
				}
			}`,
			sts: http.StatusOK,
		},
		{
			hse: true,
			pld: errorResponse,
			sts: http.StatusOK,
		},
		{
			hse: true,
			sts: http.StatusInternalServerError,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getPagesTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	rsp *Response
	tts []string
}

func (s *getPagesTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer: 		 mockTracer,
	}

	if len(s.pld) > 0 {
		s.rsp = new(Response)
		_ = json.Unmarshal([]byte(s.pld), s.rsp)
	}
}

func (s *getPagesTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getPagesTestSuite) TestGetPages() {
	pgs, err := s.clt.GetPages(s.ctx, s.dtb, s.tts)

	if s.rsp != nil && s.rsp.Query != nil {
		for _, pge := range s.rsp.Query.Pages {
			s.Assert().Equal(pgs[pge.Title], pge)
		}
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestGetPages(t *testing.T) {
	for _, testCase := range []*getPagesTestSuite{
		{
			tts: []string{"Earth", "Ninja"},
			sts: http.StatusOK,
			pld: getPagesResponse,
		},
		{
			hse: true,
			pld: errorResponse,
			sts: http.StatusOK,
		},
		{
			hse: true,
			sts: http.StatusInternalServerError,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getPageTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	err error
	dtb string
	pld string
	sts int
	rsp *Response
	ttl string
}

func (s *getPageTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer: 		 mockTracer,
	}

	if len(s.pld) > 0 {
		s.rsp = new(Response)
		_ = json.Unmarshal([]byte(s.pld), s.rsp)
	}
}

func (s *getPageTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getPageTestSuite) TestGetPage() {
	pge, err := s.clt.GetPage(s.ctx, s.dtb, s.ttl)

	if s.rsp != nil && s.rsp.Query != nil && err == nil {
		s.Assert().Equal(*pge, *s.rsp.Query.Pages[0])
	} else {
		s.Assert().Nil(pge)
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}

	if s.err != nil {
		s.Assert().Equal(err, s.err)
	}
}

func TestGetPage(t *testing.T) {
	for _, testCase := range []*getPageTestSuite{
		{
			sts: http.StatusOK,
			ttl: "Earth",
			pld: getPagesResponse,
		},
		{
			sts: http.StatusOK,
			ttl: "Not Found",
			pld: getPagesResponse,
			err: ErrPageNotFound,
			hse: true,
		},
		{
			ttl: "Earth",
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getPagesHTMLTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	dtb string
	pls []string
	tls []string
	ers []bool
	sts []int
}

func (s *getPagesHTMLTestSuite) createServer() {
	rtr := http.NewServeMux()

	for i, ttl := range s.tls {
		func(ttl string, i int) {
			url := fmt.Sprintf("/w/rest.php/v1/page/%s/html", strings.ReplaceAll(ttl, " ", "_"))

			rtr.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(s.sts[i])
				_, _ = w.Write([]byte(s.pls[i]))
			})
		}(ttl, i)
	}

	s.srv = httptest.NewServer(rtr)
}

func (s *getPagesHTMLTestSuite) SetupSuite() {
	s.createServer()

	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer: 		 mockTracer,
	}
}

func (s *getPagesHTMLTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getPagesHTMLTestSuite) TestGetPagesHTML() {
	phl := s.clt.GetPagesHTML(s.ctx, s.dtb, s.tls, 10)

	for i, ttl := range s.tls {
		phm := phl[ttl]

		if s.ers[i] {
			s.Assert().Error(phm.Error)
			s.Assert().Empty(phm.Content)
		} else {
			s.Assert().NoError(phm.Error)
			s.Assert().Equal(s.pls[i], phm.Content)
		}

		s.Assert().Equal(ttl, phm.Title)
	}

	s.Assert().Equal(len(s.tls), len(phl))
}

func TestGetPagesHTML(t *testing.T) {
	for _, testCase := range []*getPagesHTMLTestSuite{
		{
			sts: []int{http.StatusOK},
			tls: []string{"Earth"},
			pls: []string{"<p>...html goes here...</p>"},
			ers: []bool{false},
		},
		{
			sts: []int{http.StatusNotFound},
			tls: []string{"Not Found"},
			pls: []string{
				`{
					"type":"https://mediawiki.org/wiki/HyperSwitch/errors/not_found",
					"title":"Not found.",
					"method":"get",
					"detail":"Page or revision not found.",
					"uri":"/en.wikipedia.org/v1/page/html/Sadsd"
				}`,
			},
			ers: []bool{true},
		},
		{
			sts: []int{http.StatusOK, http.StatusNotFound},
			tls: []string{"Earth", "Not Found"},
			pls: []string{
				"<p>...html goes here...</p>",
				`{
					"type":"https://mediawiki.org/wiki/HyperSwitch/errors/not_found",
					"title":"Not found.",
					"method":"get",
					"detail":"Page or revision not found.",
					"uri":"/en.wikipedia.org/v1/page/html/Sadsd"
				}`,
			},
			ers: []bool{false, true},
		},
	} {
		suite.Run(t, testCase)
	}
}

type getPageHTMLTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	ttl string
}

func (s *getPageHTMLTestSuite) createServer() {
	rtr := http.NewServeMux()

	rtr.HandleFunc(fmt.Sprintf("/w/rest.php/v1/page/%s/html", strings.ReplaceAll(s.ttl, " ", "_")), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(s.sts)
		_, _ = w.Write([]byte(s.pld))
	})

	s.srv = httptest.NewServer(rtr)
}

func (s *getPageHTMLTestSuite) SetupSuite() {
	s.createServer()

	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer: 		 mockTracer,
	}
}

func (s *getPageHTMLTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getPageHTMLTestSuite) TestGetPageHTML() {
	htm, err := s.clt.GetPageHTML(s.ctx, s.dtb, s.ttl)

	if err == nil {
		s.Assert().Equal(s.pld, htm)
	} else {
		s.Assert().Empty(htm)
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestGetPageHTML(t *testing.T) {
	for _, testCase := range []*getPageHTMLTestSuite{
		{
			sts: http.StatusOK,
			ttl: "Earth",
			pld: "<p>...html goes here...</p>",
		},
		{
			sts: http.StatusFound,
			ttl: "Earth",
			pld: "<p>...html goes here...</p>",
		},
		{
			sts: http.StatusNotFound,
			ttl: "Not Found",
			pld: `{
				"type":"https://mediawiki.org/wiki/HyperSwitch/errors/not_found",
				"title":"Not found.",
				"method":"get",
				"detail":"Page or revision not found.",
				"uri":"/en.wikipedia.org/v1/page/html/Sadsd"
			}`,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getPageSummaryTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	sum string
	sts int
	ttl string
}

func (s *getPageSummaryTestSuite) createServer() {
	rtr := http.NewServeMux()

	rtr.HandleFunc(fmt.Sprintf("/api/rest_v1/page/summary/%s", strings.ReplaceAll(s.ttl, " ", "_")), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(s.sts)
		_, _ = w.Write([]byte(s.sum))
	})

	s.srv = httptest.NewServer(rtr)
}

func (s *getPageSummaryTestSuite) SetupSuite() {
	s.createServer()

	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer: 		 mockTracer,
	}
}

func (s *getPageSummaryTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getPageSummaryTestSuite) TestGetPageSummary() {
	smr, err := s.clt.GetPageSummary(s.ctx, s.dtb, s.ttl)

	if s.hse {
		s.Assert().Error(err)
		s.Assert().Nil(smr)
	} else {
		s.Assert().NoError(err)
		s.Assert().NotNil(smr)
	}
}

func TestGetPageSummary(t *testing.T) {
	for _, testCase := range []*getPageSummaryTestSuite{
		{
			sts: http.StatusOK,
			ttl: "Test",
			sum: `{"type":"disambiguation","title":"Test","displaytitle":"<span class=\"mw-page-title-main\">Test</span>","namespace":{"id":0,"text":""},"wikibase_item":"Q224615","titles":{"canonical":"Test","normalized":"Test","display":"<span class=\"mw-page-title-main\">Test</span>"},"pageid":11089416,"lang":"en","dir":"ltr","revision":"1114329112","tid":"bf9b9c00-4506-11ed-84c5-0daf85508ff9","timestamp":"2022-10-05T23:37:58Z","description":"Topics referred to by the same term","description_source":"local","content_urls":{"desktop":{"page":"https://en.wikipedia.org/wiki/Test","revisions":"https://en.wikipedia.org/wiki/Test?action=history","edit":"https://en.wikipedia.org/wiki/Test?action=edit","talk":"https://en.wikipedia.org/wiki/Talk:Test"},"mobile":{"page":"https://en.m.wikipedia.org/wiki/Test","revisions":"https://en.m.wikipedia.org/wiki/Special:History/Test","edit":"https://en.m.wikipedia.org/wiki/Test?action=edit","talk":"https://en.m.wikipedia.org/wiki/Talk:Test"}},"extract":"Test(s), testing, or TEST may refer to:Test (assessment), an educational assessment intended to measure the respondents' knowledge or other abilities","extract_html":"<p><b>Test(s)</b>, <b>testing</b>, or <b>TEST</b> may refer to:</p><ul><li>Test (assessment), an educational assessment intended to measure the respondents' knowledge or other abilities</li></ul>"}`,
		},
		{
			sts: http.StatusNotFound,
			ttl: "Not Found",
			sum: `{
				"type":"https://mediawiki.org/wiki/HyperSwitch/errors/not_found",
				"title":"Not found.",
				"method":"get",
				"detail":"Page or revision not found.",
				"uri":"/en.wikipedia.org/v1/page/summary/Sadsd"
			}`,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getLanguagesTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	lns map[string]*Language
}

func (s *getLanguagesTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer: 		 mockTracer,
	}

	s.lns = map[string]*Language{}
	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	if rsp.SiteMatrix != nil {
		for num, lng := range rsp.SiteMatrix {
			if num != "count" && num != "specials" {
				s.lns[lng.Code] = lng
			}
		}
	}
}

func (s *getLanguagesTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getLanguagesTestSuite) TestGetLanguages() {
	lns, err := s.clt.GetLanguages(s.ctx, s.dtb)

	if s.lns != nil && len(s.lns) > 0 {
		s.Assert().Len(lns, len(s.lns))

		for _, lng := range lns {
			s.Assert().Equal(s.lns[lng.Code], lng)
		}
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestGetLanguages(t *testing.T) {
	for _, testCase := range []*getLanguagesTestSuite{
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getLanguageTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	lns map[string]*Language
	err error
}

func (s *getLanguageTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: "enwiki",
		HTTPClient:      &http.Client{},
		Tracer: 		 mockTracer,
	}

	s.lns = map[string]*Language{}
	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	if rsp.SiteMatrix != nil {
		for num, lng := range rsp.SiteMatrix {
			if num != "count" && num != "specials" {
				s.lns[lng.Code] = lng
			}
		}
	}
}

func (s *getLanguageTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getLanguageTestSuite) TestGetLanguage() {
	lng, err := s.clt.GetLanguage(s.ctx, s.dtb)

	if s.hse {
		s.Assert().Nil(lng)
		s.Assert().Error(err)
	} else {
		s.Assert().Equal(s.lns[lng.Code], lng)
		s.Assert().NoError(err)
	}

	if s.err != nil {
		s.Assert().Equal(s.err, err)
	}
}

func TestGetLanguage(t *testing.T) {
	for _, testCase := range []*getLanguageTestSuite{
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
			dtb: "afwikibooks",
		},
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
			hse: true,
			dtb: "not found",
			err: ErrLanguageNotFound,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getProjectTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	prs map[string]*Project
	err error
}

func (s *getProjectTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: "enwiki",
		HTTPClient:      &http.Client{},
		Tracer: 		 mockTracer,
	}

	s.prs = map[string]*Project{}
	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	if rsp.SiteMatrix != nil {
		for num, lng := range rsp.SiteMatrix {
			if num != "count" && num != "specials" {
				for _, prj := range lng.Projects {
					s.prs[prj.DBName] = prj
				}
			}
		}
	}
}

func (s *getProjectTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getProjectTestSuite) TestGetProject() {
	prj, err := s.clt.GetProject(s.ctx, s.dtb)

	if s.hse {
		s.Assert().Nil(prj)
		s.Assert().Error(err)
	} else {
		s.Assert().Equal(s.prs[prj.DBName], prj)
		s.Assert().NoError(err)
	}

	if s.err != nil {
		s.Assert().Equal(s.err, err)
	}
}

func TestGetProject(t *testing.T) {
	for _, testCase := range []*getProjectTestSuite{
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
			dtb: "afwikibooks",
		},
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
			hse: true,
			dtb: "not found",
			err: ErrProjectNotFound,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getProjectsTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	hse bool
	dtb string
	pld string
	sts int
	prs []*Project
	err error
}

func (s *getProjectsTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: "enwiki",
		HTTPClient:      &http.Client{},
		Tracer: 		 mockTracer,
	}

	s.prs = []*Project{}
	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	if rsp.SiteMatrix != nil {
		for num, lng := range rsp.SiteMatrix {
			if num != "count" && num != "specials" {
				s.prs = append(s.prs, lng.Projects...)
			}
		}
	}
}

func (s *getProjectsTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getProjectsTestSuite) TestGetProjects() {
	prl, err := s.clt.GetProjects(s.ctx, s.dtb)

	if s.hse {
		s.Assert().Nil(prl)
		s.Assert().Error(err)
	} else {
		s.Assert().Equal(len(s.prs), len(prl))
		s.Assert().NoError(err)
	}

	if s.err != nil {
		s.Assert().Equal(s.err, err)
	}
}

func TestGetProjects(t *testing.T) {
	for _, testCase := range []*getProjectsTestSuite{
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
			dtb: "enwiki",
		},
		{
			sts: http.StatusOK,
			pld: getLanguagesResponse,
			hse: true,
			dtb: "not found",
			err: ErrProjectNotFound,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getNamespacesTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	rsp *Response
	hse bool
	dtb string
	pld string
	sts int
}

func (s *getNamespacesTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer: 		 mockTracer,
	}

	s.rsp = new(Response)
	_ = json.Unmarshal([]byte(s.pld), s.rsp)
}

func (s *getNamespacesTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getNamespacesTestSuite) TestGetNamespaces() {
	nms, err := s.clt.GetNamespaces(s.ctx, s.dtb)

	if s.rsp.Query != nil && len(s.rsp.Query.Namespaces) > 0 {
		s.Assert().Len(nms, len(s.rsp.Query.Namespaces))

		for _, nsp := range s.rsp.Query.Namespaces {
			s.Assert().Equal(s.rsp.Query.Namespaces[nsp.ID], nsp)
		}
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestGetNamespaces(t *testing.T) {
	for _, testCase := range []*getNamespacesTestSuite{
		{
			sts: http.StatusOK,
			pld: `{"batchcomplete":true,"query":{"namespaces":{"-2":{"id":-2,"case":"first-letter","name":"Media","subpages":false,"canonical":"Media","content":false,"nonincludable":false},"-1":{"id":-1,"case":"first-letter","name":"Special","subpages":false,"canonical":"Special","content":false,"nonincludable":false},"0":{"id":0,"case":"first-letter","name":"","subpages":false,"content":true,"nonincludable":false},"1":{"id":1,"case":"first-letter","name":"Talk","subpages":true,"canonical":"Talk","content":false,"nonincludable":false},"2":{"id":2,"case":"first-letter","name":"User","subpages":true,"canonical":"User","content":false,"nonincludable":false},"3":{"id":3,"case":"first-letter","name":"User talk","subpages":true,"canonical":"User talk","content":false,"nonincludable":false},"4":{"id":4,"case":"first-letter","name":"Wikipedia","subpages":true,"canonical":"Project","content":false,"nonincludable":false},"5":{"id":5,"case":"first-letter","name":"Wikipedia talk","subpages":true,"canonical":"Project talk","content":false,"nonincludable":false},"6":{"id":6,"case":"first-letter","name":"File","subpages":false,"canonical":"File","content":false,"nonincludable":false},"7":{"id":7,"case":"first-letter","name":"File talk","subpages":true,"canonical":"File talk","content":false,"nonincludable":false},"8":{"id":8,"case":"first-letter","name":"MediaWiki","subpages":false,"canonical":"MediaWiki","content":false,"nonincludable":false,"namespaceprotection":"editinterface"},"9":{"id":9,"case":"first-letter","name":"MediaWiki talk","subpages":true,"canonical":"MediaWiki talk","content":false,"nonincludable":false},"10":{"id":10,"case":"first-letter","name":"Template","subpages":true,"canonical":"Template","content":false,"nonincludable":false},"11":{"id":11,"case":"first-letter","name":"Template talk","subpages":true,"canonical":"Template talk","content":false,"nonincludable":false},"12":{"id":12,"case":"first-letter","name":"Help","subpages":true,"canonical":"Help","content":false,"nonincludable":false},"13":{"id":13,"case":"first-letter","name":"Help talk","subpages":true,"canonical":"Help talk","content":false,"nonincludable":false},"14":{"id":14,"case":"first-letter","name":"Category","subpages":true,"canonical":"Category","content":false,"nonincludable":false},"15":{"id":15,"case":"first-letter","name":"Category talk","subpages":true,"canonical":"Category talk","content":false,"nonincludable":false},"100":{"id":100,"case":"first-letter","name":"Portal","subpages":true,"canonical":"Portal","content":false,"nonincludable":false},"101":{"id":101,"case":"first-letter","name":"Portal talk","subpages":true,"canonical":"Portal talk","content":false,"nonincludable":false},"118":{"id":118,"case":"first-letter","name":"Draft","subpages":true,"canonical":"Draft","content":false,"nonincludable":false},"119":{"id":119,"case":"first-letter","name":"Draft talk","subpages":true,"canonical":"Draft talk","content":false,"nonincludable":false},"710":{"id":710,"case":"first-letter","name":"TimedText","subpages":false,"canonical":"TimedText","content":false,"nonincludable":false},"711":{"id":711,"case":"first-letter","name":"TimedText talk","subpages":false,"canonical":"TimedText talk","content":false,"nonincludable":false},"828":{"id":828,"case":"first-letter","name":"Module","subpages":true,"canonical":"Module","content":false,"nonincludable":false},"829":{"id":829,"case":"first-letter","name":"Module talk","subpages":true,"canonical":"Module talk","content":false,"nonincludable":false},"2300":{"id":2300,"case":"case-sensitive","name":"Gadget","subpages":false,"canonical":"Gadget","content":false,"nonincludable":false,"namespaceprotection":"gadgets-edit"},"2301":{"id":2301,"case":"case-sensitive","name":"Gadget talk","subpages":false,"canonical":"Gadget talk","content":false,"nonincludable":false},"2302":{"id":2302,"case":"case-sensitive","name":"Gadget definition","subpages":false,"canonical":"Gadget definition","content":false,"nonincludable":false,"namespaceprotection":"gadgets-definition-edit","defaultcontentmodel":"GadgetDefinition"},"2303":{"id":2303,"case":"case-sensitive","name":"Gadget definition talk","subpages":false,"canonical":"Gadget definition talk","content":false,"nonincludable":false}}}}`,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getRandomPagesTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	rsp *Response
	hse bool
	dtb string
	pld string
	sts int
}

func (s *getRandomPagesTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer: 		 mockTracer,
	}

	s.rsp = new(Response)
	_ = json.Unmarshal([]byte(s.pld), s.rsp)
}

func (s *getRandomPagesTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getRandomPagesTestSuite) TestGetRandomPages() {
	pgs, err := s.clt.GetRandomPages(s.ctx, s.dtb)

	if s.rsp.Query != nil && len(s.rsp.Query.Random) > 0 {
		s.Assert().Len(pgs, len(s.rsp.Query.Random))

		itl := []string{}
		rtl := []string{}

		for _, rnd := range s.rsp.Query.Random {
			itl = append(itl, rnd.Title)
		}

		for _, pge := range pgs {
			rtl = append(rtl, pge.Title)
		}

		s.Assert().ElementsMatch(itl, rtl)
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestGetRandomPages(t *testing.T) {
	for _, testCase := range []*getRandomPagesTestSuite{
		{
			sts: http.StatusOK,
			pld: `{
				"batchcomplete": true,
				"continue": {
				  "rncontinue": "0.626892872077|0.626893505289|40964313|0",
				  "continue": "-||"
				},
				"query": {
				  "random": [
					{
					  "id": 823840,
					  "ns": 0,
					  "title": "Nintendo Power"
					},
					{
					  "id": 60232671,
					  "ns": 0,
					  "title": "Umenomoto Station"
					},
					{
					  "id": 17902603,
					  "ns": 0,
					  "title": "2008 ANAPROF Clausura"
					},
					{
					  "id": 5101698,
					  "ns": 0,
					  "title": "Siege of Otate"
					},
					{
					  "id": 27747856,
					  "ns": 0,
					  "title": "William J. Hughes Technical Center"
					}
				  ]
				}
			  }`,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getUsersTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	urs map[int]*User
	hse bool
	dtb string
	ids []int
	pld string
	sts int
}

func (s *getUsersTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer: 		 mockTracer,
	}

	s.urs = map[int]*User{}
	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	if rsp.Query != nil {
		for _, usr := range rsp.Query.Users {
			s.urs[usr.UserID] = usr
		}
	}
}

func (s *getUsersTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getUsersTestSuite) TestGetUsers() {
	urs, err := s.clt.GetUsers(s.ctx, s.dtb, s.ids)

	if s.urs != nil && len(s.urs) > 0 {
		s.Assert().Len(urs, len(s.urs))

		for _, usr := range urs {
			s.Assert().Equal(s.urs[usr.UserID], usr)
		}
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestGetUsers(t *testing.T) {
	for _, testCase := range []*getUsersTestSuite{
		{
			ids: []int{2, 3, 5},
			sts: http.StatusOK,
			pld: getUsersResponse,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type getUserTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	err error
	urs map[int]*User
	hse bool
	dtb string
	uid int
	pld string
	sts int
}

func (s *getUserTestSuite) SetupSuite() {
	s.srv = createActionsAPIServer(s.sts, s.pld)
	s.dtb = "enwiki"
	s.ctx = context.Background()
	s.clt = &Client{
		DefaultURL:      s.srv.URL,
		DefaultDatabase: s.dtb,
		HTTPClient:      &http.Client{},
		Tracer: 		 mockTracer,
	}

	s.urs = map[int]*User{}
	rsp := new(Response)
	_ = json.Unmarshal([]byte(s.pld), rsp)

	if rsp.Query != nil {
		for _, usr := range rsp.Query.Users {
			s.urs[usr.UserID] = usr
		}
	}
}

func (s *getUserTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getUserTestSuite) TestGetUser() {
	usr, err := s.clt.GetUser(s.ctx, s.dtb, s.uid)

	if s.urs != nil && len(s.urs) > 0 {
		s.Assert().Equal(s.urs[usr.UserID], usr)
	}

	if s.hse {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}

	if s.err != nil {
		s.Assert().Equal(s.err, err)
	}
}

func TestGetUser(t *testing.T) {
	for _, testCase := range []*getUserTestSuite{
		{
			uid: 5,
			sts: http.StatusOK,
			pld: getUsersResponse,
		},
		{
			uid: 10,
			sts: http.StatusOK,
			pld: getPagesResponse,
			err: ErrUserNotFound,
			hse: true,
		},
		{
			sts: http.StatusOK,
			pld: errorResponse,
			hse: true,
		},
		{
			sts: http.StatusInternalServerError,
			hse: true,
		},
	} {
		suite.Run(t, testCase)
	}
}

type retryAfterTestSuite struct {
	suite.Suite
	ctx context.Context
	srv *httptest.Server
	clt *Client
	rtv int
	era bool
	hre bool
}

func (s *retryAfterTestSuite) SetupSuite() {
	rtr := http.NewServeMux()
	cnt := 2

	rtr.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if cnt > 0 {
			w.Header().Add("Retry-After", "Wed, 21 Oct 2015 07:28:00 GMT")
			w.Header().Add("Retry-After", strconv.Itoa(s.rtv))
			w.WriteHeader(http.StatusTooManyRequests)
		}

		cnt--
	})

	s.srv = httptest.NewServer(rtr)
	s.ctx = context.Background()
	s.clt = &Client{
		HTTPClient:       &http.Client{},
		DefaultURL:       s.srv.URL,
		DefaultDatabase:  "enwiki",
		EnableRetryAfter: s.era,
		Tracer: 		  mockTracer,
	}
}

func (s *retryAfterTestSuite) TestRetryAfrer() {
	_, err := s.clt.GetLanguages(s.ctx, s.clt.DefaultDatabase)

	if s.hre {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}

	_, err = s.clt.GetPageHTML(s.ctx, s.clt.DefaultDatabase, "Earth")

	if s.hre {
		s.Assert().Error(err)
	} else {
		s.Assert().NoError(err)
	}
}

func TestRetryAfter(t *testing.T) {
	for _, testCase := range []*retryAfterTestSuite{
		{
			era: true,
			hre: false,
			rtv: 1,
		},
		{
			era: false,
			hre: true,
			rtv: 1,
		},
	} {
		suite.Run(t, testCase)
	}
}

func createLiftWingAPIServer(sts int, pld string) *httptest.Server {
	rtr := http.NewServeMux()

	rtr.HandleFunc("/service/lw/inference/v1/models/revertrisk-language-agnostic:predict", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(sts)
		if sts == http.StatusOK {
			_, _ = w.Write([]byte(pld))
		} else {
			_, _ = w.Write([]byte("invalid status code"))
		}
	})

	return httptest.NewServer(rtr)
}

type getScoreTestSuite struct {
	suite.Suite
	ctx context.Context
	clt *Client
	srv *httptest.Server
	sts int    // status codes
	pld string // payload
	rid int
	lng string
	prj string
	mdl string
	err error
}

func (s *getScoreTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.pld = strings.ReplaceAll(s.pld, "\n", "")
	s.pld = strings.ReplaceAll(s.pld, " ", "")
	s.srv = createLiftWingAPIServer(s.sts, s.pld)
	s.prj = ""
	s.clt = &Client{
		HTTPClientLiftWing: &http.Client{},
		LiftWingBaseURL:    s.srv.URL + "/service/lw/inference/v1/models/",
		Tracer: 			mockTracer,
	}
}

func (s *getScoreTestSuite) TearDownSuite() {
	s.srv.Close()
}

func (s *getScoreTestSuite) TestGetScore() {
	rsp, err := s.clt.GetScore(s.ctx, s.rid, s.lng, s.prj, s.mdl)

	if s.err != nil {
		s.Assert().Error(err)
		s.Assert().Contains(err.Error(), s.err.Error())
	} else {
		s.Assert().NoError(err)
		rrk, _ := json.Marshal(rsp)
		s.Assert().Equal(s.pld, string(rrk))
	}
}

func TestGetScore(t *testing.T) {
	for _, testCase := range []*getScoreTestSuite{
		{
			sts: http.StatusOK,
			pld: liftWingPayload,
			mdl: "revertrisk",
			lng: "en",
			err: nil,
		},
		{
			sts: http.StatusBadRequest,
			pld: liftWingPayload,
			mdl: "revertrisk",
			lng: "en",
			err: errors.New("400 Bad Request:invalidstatuscode"),
		},
	} {
		suite.Run(t, testCase)
	}
}
