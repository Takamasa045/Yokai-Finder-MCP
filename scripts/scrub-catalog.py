#!/usr/bin/env python3
"""Strip generation residue from the yokai catalog and add rank-1 Japanese lore."""

from __future__ import annotations

import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DATA = ROOT / "internal" / "yokai" / "data"

NEEDLES = [
    "do not merge",
    "as its own plate",
    "encyclopedia plate",
    "encyclopedia body",
    "folk plate",
    "the workplace is",
    "neighbour card",
    "neighbor card",
    "well card",
    "faced shell",
    "別項",
    "巨大側",
    "図鑑に残る",
]

PROFILE_ORDER = [
    "name",
    "nativeName",
    "region",
    "category",
    "categoryEn",
    "summary",
    "summaryJa",
    "legends",
    "legendsJa",
    "traits",
    "traitsJa",
    "motifs",
    "motifsJa",
    "funFact",
    "funFactJa",
    "searchQuery",
    "creativeHooks",
    "creativeHooksJa",
    "sources",
]

INDEX_ORDER = [
    "name",
    "nativeName",
    "category",
    "region",
    "blurbJa",
    "tags",
    "tone",
    "famousRank",
    "aliases",
]

CAT_REGION = {
    "水系": "川・沼",
    "海": "海",
    "山系": "山中",
    "屋敷": "家屋",
    "町": "町",
    "道中": "街道",
    "田畑": "田畑",
    "霊火": "夜空",
    "現代伝承": "都市",
    "気象": "空・天候",
    "鬼": "山里",
    "狐狸": "里・社",
    "付喪神": "家の道具",
    "死霊": "墓・家",
    "変化": "里",
    "異形": "古典",
    "古典": "古典",
    "憑きもの": "家筋",
    "予言": "海辺",
}

CAT_EN = {
    "水系": "water yokai",
    "海": "sea yokai",
    "山系": "mountain yokai",
    "屋敷": "house yokai",
    "町": "town yokai",
    "道中": "road yokai",
    "田畑": "field yokai",
    "霊火": "ghost-fire yokai",
    "現代伝承": "modern urban legend",
    "気象": "weather yokai",
    "鬼": "oni",
    "狐狸": "fox or tanuki spirit",
    "付喪神": "tsukumogami",
    "死霊": "ghost",
    "変化": "shapeshifter",
    "異形": "chimeric yokai",
    "古典": "classical yokai",
    "憑きもの": "possessing spirit",
    "予言": "omen yokai",
}

PLACE_HINTS = (
    "全国", "東北", "関東", "関西", "九州", "四国", "北海道", "沖縄", "江戸", "京都",
    "奈良", "大阪", "東京", "熊本", "秋田", "岩手", "海", "山", "川", "沼", "森",
    "社", "寺", "城", "村", "町", "国", "県", "道", "里", "島", "湖", "峠", "野",
)

RANK1 = {
    "河童": {
        "summaryJa": "川や沼に棲む、亀の甲羅を背負った悪戯好きの水の妖怪。頭の皿の水が力の源で、キュウリと相撲が大好物。",
        "legendsJa": [
            "川辺で旅人に相撲を挑み、負けた者の尻子玉を抜くという。",
            "深くお辞儀を返すうちに頭の皿の水をこぼし、力を失う。",
            "キュウリを供えると、代わりに災いを避けられると村では語る。",
        ],
        "traitsJa": ["甲羅", "水かき", "頭の皿", "キュウリ好き"],
        "creativeHooksJa": [
            "川の水利権を河童と書面で結ぶ村の話。",
            "相撲の礼が勝負を決める場面。",
            "皿の水を守るための笠を贈る子ども。",
        ],
    },
    "天狗": {
        "summaryJa": "高い鼻と翼を持つ山の護り手。慢心した旅人や僧を懲らしめ、見込んだ者には剣術を授けるという。",
        "legendsJa": [
            "慢心した僧をさらい、山中で修行させてから返す天狗隠しの話がある。",
            "見込んだ弟子には剣術と羽団扇の風を教える。",
            "修験の山では、天狗は神仏の使いとしても語られる。",
        ],
        "traitsJa": ["長い鼻", "翼", "羽団扇", "山の矜持"],
        "creativeHooksJa": [
            "峠の吊り橋で天狗が礼を試す。",
            "杉の奥に隠れた天狗の道場。",
            "山そのものが師匠である剣の話。",
        ],
    },
    "雪女": {
        "summaryJa": "吹雪の夜に現れる氷のように美しい女性の妖怪。慈悲と冷徹さをあわせ持ち、心優しい者を見逃すこともある。",
        "legendsJa": [
            "温かさを分けた樵を助け、夜明けとともに雪へ還る。",
            "出会ったことを口外しない誓いを破ると、妻の姿が雪に崩れる。",
            "氷の赤子を抱いて現れ、旅人の情けを試す話もある。",
        ],
        "traitsJa": ["白い着物", "冷気の息", "足音のない歩み", "吹雪の体"],
        "creativeHooksJa": [
            "春との休戦を雪女が交渉する。",
            "家庭の暖かさが誓いそのものである家。",
            "吹雪の夜にだけ開く宿。",
        ],
    },
    "九尾の狐": {
        "summaryJa": "変化と幻術を操る狐の霊。知恵と悪戯の両方を持ち、稲荷の使いとも、朝廷を揺るがす妖狐とも語られる。",
        "legendsJa": [
            "人に化けて里に交わり、灯火や吉兆で稲荷と結び付けられる。",
            "玉藻前の物語では、九尾の狐が妃に化けて帝を病ませたという。",
            "尾の数が智恵と術の深さを示す、と絵巻は描く。",
        ],
        "traitsJa": ["変化", "狐火", "複数の尾", "幻術"],
        "creativeHooksJa": [
            "社の使いと妖狐の境界が揺らぐ夜。",
            "鏡にだけ九つの尾が映る人物。",
            "約束の札を食べる狐。",
        ],
    },
    "化け猫": {
        "summaryJa": "長く飼われた猫が尾を叉に分け、人を恨んで化けるという妖怪。夜の座敷で灯をなめ、仇を取る話が多い。",
        "legendsJa": [
            "虐待された猫が女に化け、仇の家を滅ぼす。",
            "二股の尾を隠して侍の側に仕える。",
            "行灯の油をなめる姿を見られると、正体が知れる。",
        ],
        "traitsJa": ["二股の尾", "人化", "復讐", "夜の灯"],
        "creativeHooksJa": [
            "飼い主の罪を猫だけが覚えている家。",
            "油の減り方で正体がバレる旅館。",
            "仇討ちをやめた化け猫の余生。",
        ],
    },
    "海坊主": {
        "summaryJa": "沖に現れる巨大な坊主頭の海の妖怪。船を覆うほど大きく、沈黙と引き換えに通航を許すともいう。",
        "legendsJa": [
            "凪の夜、船体ほどある黒い頭が海面を割る。",
            "声を出すと沈む、と漁師は子に教える。",
            "樽や水を求めることがあり、拒むと波が荒れる。",
        ],
        "traitsJa": ["坊主頭", "巨大", "無言", "船幽霊に近い"],
        "creativeHooksJa": [
            "海坊主と取引する無線。",
            "頭だけが港の防波堤になる夜。",
            "沈黙を守った船だけが帰ってくる。",
        ],
    },
    "餓者髑髏": {
        "summaryJa": "飢饉や戦で死んだ者の怨みが集まった巨大な骸骨。夜、野を歩き、人を掴んで噛むという。",
        "legendsJa": [
            "歯の間から血の雨が降る、と絵草紙は脅す。",
            "個々の霊ではなく、飢えそのものが骨の形を取る。",
            "供養の灯が遠い村ほど、出やすいとされる。",
        ],
        "traitsJa": ["巨大な骸骨", "飢え", "夜歩き", "歯"],
        "creativeHooksJa": [
            "飢饉アーカイブが骨になる図書館。",
            "歯の一本が村人の名前を覚えている。",
            "供養の行列と骸骨がすれ違う夜。",
        ],
    },
    "アマビエ": {
        "summaryJa": "肥後の海に現れたという、鳥のくちばしを持つ予言の妖怪。疫病の流行を告げ、自分の絵を写せば難を避けると教えた。",
        "legendsJa": [
            "光る海から現れ、役人と対話して去る。",
            "「私の姿を写して広めるがよい」と疫病除けを託す。",
            "近世の瓦版が、その図像を全国へ運んだ。",
        ],
        "traitsJa": ["くちばし", "鱗", "予言", "絵札"],
        "creativeHooksJa": [
            "写した絵が現代の画面で増殖する。",
            "アマビエが最新の病名を知っている。",
            "海辺の役所に残る一枚の瓦版。",
        ],
    },
    "座敷童子": {
        "summaryJa": "家に宿る童子の霊。いる家は栄え、去ると衰えるという。遠野を中心に、東北の屋敷譚として語られる。",
        "legendsJa": [
            "赤い顔の童が座敷で鞠をつく音がする。",
            "丁寧に遇した家は富み、追い出された家は傾く。",
            "去る夜、誰もいない廊下に小さな足跡が残る。",
        ],
        "traitsJa": ["童の姿", "家の運", "隠れ居", "東北"],
        "creativeHooksJa": [
            "空き家バンクにだけ現れる童。",
            "去る条件を契約書にした家族。",
            "観光の前に、座敷の音を聞く。",
        ],
    },
    "のっぺらぼう": {
        "summaryJa": "目も鼻もない平らな顔の妖怪。深夜の道で人を呼び止め、自分の顔を見せて言葉を奪う。",
        "legendsJa": [
            "助けを求めた女の顔が、のっぺらりと消える。",
            "逃げ込んだ茶店の店主も同じ顔だった、という落ちが定番。",
            "顔のない者は、名前で呼ばれることを嫌うともいう。",
        ],
        "traitsJa": ["平滑な顔", "夜道", "驚き", "模倣"],
        "creativeHooksJa": [
            "身分証の顔写真が空白の人物。",
            "のっぺらぼうが顔を借りて一日だけ働く。",
            "鏡だけが、かつての五官を覚えている。",
        ],
    },
    "提灯お化け": {
        "summaryJa": "捨てられ、破れた提灯に霊が宿った付喪神。裂けた紙が口になり、長い舌を出して人を脅す。",
        "legendsJa": [
            "夜道の提灯が急に消え、次の瞬間に顔になる。",
            "百年近い道具ほど、付喪神になりやすいとされる。",
            "新しい火を入れ直すと、静かになることもある。",
        ],
        "traitsJa": ["破れた紙", "長い舌", "一本足", "夜の灯"],
        "creativeHooksJa": [
            "LEDに生まれ変わっても口だけ残る提灯。",
            "祭の列から一基だけが笑う。",
            "火を分けてもらう代わりに道案内をする。",
        ],
    },
    "唐傘お化け": {
        "summaryJa": "古い唐傘が跳ねる付喪神。一本の脚と一つの目、長い舌が特徴で、雨の日に人を驚かせる。",
        "legendsJa": [
            "雨宿りの軒で、置いてあった傘がひとりで立つ。",
            "目が覚めた傘は、閉じられることを嫌う。",
            "鳥山石燕の絵が、この姿を広く知らしめた。",
        ],
        "traitsJa": ["一本足", "一つ目", "唐傘", "雨"],
        "creativeHooksJa": [
            "置き傘が客を選ぶ喫茶店。",
            "台風の夜にだけ行列する傘。",
            "骨が折れた傘を修理する付喪神。",
        ],
    },
    "化け狸": {
        "summaryJa": "人や物に化ける狸の妖怪。腹鼓で幻を見せ、葉を銭に変える。四国や佐渡など、土地ごとの名君・大狸としても語られる。",
        "legendsJa": [
            "葉を貨幣に見せて酒を飲むが、朝には葉に戻る。",
            "大名行列や寺を丸ごと幻で作る大狸の話がある。",
            "化け負けすると、腹を抱えて逃げる。",
        ],
        "traitsJa": ["腹鼓", "葉の銭", "幻術", "茶目"],
        "creativeHooksJa": [
            "キャッシュレス決済に負ける狸。",
            "幻の駅が時刻表に一日だけ載る。",
            "化け比べを観光資源にした町。",
        ],
    },
    "猫又": {
        "summaryJa": "老いた猫が尾を二つに分け、人語を話し、時に死者を操るという。化け猫より山と長寿のイメージが強い。",
        "legendsJa": [
            "山中で人を呼び、足跡だけが二つに分かれる。",
            "踊る猫又の宴に招かれると、夜明けまで家に帰れない。",
            "十七歳を超えた猫は、座敷から遠ざける地方もある。",
        ],
        "traitsJa": ["二尾", "人語", "長寿", "山"],
        "creativeHooksJa": [
            "十七歳の猫の誕生日に家族が会議する。",
            "猫又だけが読める過去帳。",
            "山道の分岐が尾のように分かれる。",
        ],
    },
    "木霊": {
        "summaryJa": "木に宿る霊。山彦の音とも、伐採を戒める木の神とも語られる。古い木を切ると、声や病で返すという。",
        "legendsJa": [
            "斧を入れる前に、木へ声をかける習わしがある。",
            "切り倒した夜から、家の中で木の匂いがする。",
            "山で呼んだ声が、少し遅れて自分の声で返る。",
        ],
        "traitsJa": ["木の霊", "反響", "伐採禁忌", "長寿"],
        "creativeHooksJa": [
            "再開発の説明会に木霊が傍聴する。",
            "年輪が日記になっている巨木。",
            "声の遅延だけが通信手段の村。",
        ],
    },
    "山姥": {
        "summaryJa": "山に住む老女の妖怪。旅人をもてなし、夜になると正体を現して食らう話と、金太郎を育てた聖母のような話が共存する。",
        "legendsJa": [
            "峠の小屋で一夜の宿を貸し、深夜に巨大化する。",
            "逃げた旅人が、後ろで伸びる山姥の腕を見る。",
            "一方で、英雄を育てる山の母としても語られる。",
        ],
        "traitsJa": ["山の小屋", "伸縮する体", "両義性", "髪"],
        "creativeHooksJa": [
            "民宿レビューが極端に分かれる山。",
            "育てた子が里の役人になる。",
            "腕だけが谷を越えて門を開ける。",
        ],
    },
    "鬼": {
        "summaryJa": "角と虎の皮を持つ異形。人の罪や疫病の象徴でもあり、節分で追われる一方、門を守る鬼もある。個別の名を持つ鬼は別に語られる。",
        "legendsJa": [
            "紅と藍の鬼が、地獄の庁で亡者を裁く絵が多い。",
            "節分の豆は、鬼を追うと同時に福を招く。",
            "人だった者が鬼になる話と、山の鬼が里へ下りる話がある。",
        ],
        "traitsJa": ["角", "金棒", "虎の皮", "紅藍"],
        "creativeHooksJa": [
            "鬼が人事評価をする役所。",
            "豆を数え直す節分。",
            "角を折った鬼の転職。",
        ],
    },
    "酒呑童子": {
        "summaryJa": "大江山に拠った鬼の頭領。都から人をさらい、源頼光らに討たれた酒飲みの童子姿の鬼として語られる。",
        "legendsJa": [
            "美女に化けた鬼が、都の貴族を山へ誘う。",
            "神酒で酔ったすきに首を討たれる。",
            "落ちた首が頼光の兜に噛みついた、という場面が残る。",
        ],
        "traitsJa": ["酒", "童子の顔", "大江山", "首"],
        "creativeHooksJa": [
            "禁酒した酒呑童子の二日目。",
            "首だけが観光案内をする山。",
            "鬼の宴会を再現する料理店の禁忌。",
        ],
    },
    "ぬらりひょん": {
        "summaryJa": "夜の家に上がり込み、主のように振る舞う滑（ぬめ）った老人の妖怪。正体がつかめても、なぜか追い出しにくい。",
        "legendsJa": [
            "湯を使い、煙草を吸い、家人を客のように扱う。",
            "顔を覚えようとすると、特徴が滑り落ちる。",
            "百鬼夜行の指揮者に擬せられることもある。",
        ],
        "traitsJa": ["頭の皿のような頭", "厚かましさ", "滑る記憶", "夜の訪問"],
        "creativeHooksJa": [
            "住民票に先に名前がある訪問者。",
            "会議の司会がぬらりひょん。",
            "追い出す作法だけが家訓の一族。",
        ],
    },
    "絡新婦": {
        "summaryJa": "美しい女に化けて人を誘う蜘蛛の妖怪。糸で部屋を張り、獲物を吊るす。山寺や古井戸の話が多い。",
        "legendsJa": [
            "琴を弾く女の後ろに、糸が見える。",
            "武士が灯で糸を焼き、巨大な蜘蛛を討つ。",
            "子蜘蛛が散って、次の年の災いになる",
        ],
        "traitsJa": ["女の姿", "糸", "毒", "天井"],
        "creativeHooksJa": [
            "Wi-Fiの線が糸になる古民家。",
            "琴の弦だけが獲物を覚えている。",
            "蜘蛛と共存する養蚕の村。",
        ],
    },
    "玉藻前": {
        "summaryJa": "鳥羽院に仕えた美女で、その正体は九尾の狐と語られる。金の毛の狐として暴れたのち、那須の殺生石になったという。",
        "legendsJa": [
            "病の帝のそばに侍り、陰陽師に正体を暴かれる。",
            "追われて那須野で石となり、触れるものを殺す。",
            "狐の智恵は、宮廷の学問と見分けがつかない。",
        ],
        "traitsJa": ["宮廷の美", "九尾", "殺生石", "変化"],
        "creativeHooksJa": [
            "石の放射線測定が物語と重なる。",
            "玉藻前の日記だけが残る文庫。",
            "狐が人間の病を治してしまう逆説。",
        ],
    },
    "清姫": {
        "summaryJa": "安珍を追って日高川を渡り、道成寺で鐘に籠もる僧を焼き殺した女。嫉妬が蛇体へ化する物語として語られる。",
        "legendsJa": [
            "愛を拒まれた清姫が、川を蛇になって渡る。",
            "鐘の中の安珍は、熱で溶けたように死ぬ。",
            "道成寺の能は、その執心をいまも舞う。",
        ],
        "traitsJa": ["執心", "蛇体", "川", "鐘"],
        "creativeHooksJa": [
            "鐘の鋳造記録に残る温度。",
            "清姫が安珍を許す別の結末。",
            "川の水位が嫉妬と連動する。",
        ],
    },
    "口裂け女": {
        "summaryJa": "大きなマスクの下に裂ける口を持つ、現代の都市伝説。子どもに「私、きれい？」と問い、答えによって運命が変わる。",
        "legendsJa": [
            "放課後の道に、長い髪の女が待つ。",
            "「普通」と答えると追われ、「きれい」でも裂かれた口を見せられる、と語りは分岐する。",
            "昭和後期に学校を中心へ広がり、その時代ごとの顔を持つ。",
        ],
        "traitsJa": ["マスク", "裂ける口", "問い", "放課後"],
        "creativeHooksJa": [
            "アンケートの設問が命取りになる。",
            "マスク文化のあとに現れる新しい口。",
            "答えを保留した子どもだけが帰る。",
        ],
    },
    "人魚": {
        "summaryJa": "人の頭と魚の体を持つ海の存在。肉を食べると不老になる八百比丘尼の話と、漂着して凶兆となる話が並ぶ。",
        "legendsJa": [
            "若狭で人魚の肉を食べた少女が、八百歳まで生きたという。",
            "人魚の漂着は、飢饉や乱の前触れとされる地方がある。",
            "西洋の人魚姫とは違い、日本では瑞祥と怪異のあいだにいる。",
        ],
        "traitsJa": ["人面魚体", "不老の肉", "漂着", "海"],
        "creativeHooksJa": [
            "賞味期限のない肉を拒む村。",
            "人魚が自分の肉の伝説を嫌う。",
            "漂着記録だけが残る漁協の倉庫。",
        ],
    },
    "付喪神": {
        "summaryJa": "長く使われた道具に霊が宿ったものの総称。九十九年で神になる、とも、捨てられた恨みで化ける、ともいう。傘や提灯は個別に名を持つ。",
        "legendsJa": [
            "道具は歳月とともに、人の気配を覚える。",
            "煤払いで捨てられたものが、百鬼の列に加わる絵がある。",
            "丁寧に供養された道具は、静かに役目を終える。",
        ],
        "traitsJa": ["道具の霊", "歳月", "廃棄", "百鬼"],
        "creativeHooksJa": [
            "サブスクで歳を取れない道具たち。",
            "修理屋が付喪神の戸籍を預かる。",
            "捨てる前の挨拶が物語の核になる。",
        ],
    },
    "幽霊": {
        "summaryJa": "死んでもこの世に残る人の気配。足がなく、恨みや未練で現れることが多い。個別の名を持つ霊は、それぞれ別の物語を持つ。",
        "legendsJa": [
            "盆に家へ帰る霊と、成仏できない霊は、絵の足の描き方が違う。",
            "恨みの霊は、相手の名を呼び続ける。",
            "線香と水は、今も基本のもてなしである。",
        ],
        "traitsJa": ["足がない", "未練", "白い装束", "夜"],
        "creativeHooksJa": [
            "幽霊が行政手続きに遅れる。",
            "足のない者が踏む踏切。",
            "未練の内容が、遺産分割協議書になっている。",
        ],
    },
    "トイレの花子さん": {
        "summaryJa": "学校の便所に棲む少女の霊。三つ目の個室をノックすると応える、という現代の学校怪談の代表。",
        "legendsJa": [
            "赤い服の花子が、呼び出した子に手を引く。",
            "学校が変わっても、三番目の個室という設定は残りやすい。",
            "花子さんは、いじめや事故の記憶と結び付けて語られることもある。",
        ],
        "traitsJa": ["学校", "便所", "少女", "呼び出し"],
        "creativeHooksJa": [
            "廃校の個室にだけ電波がある。",
            "花子さんが不在届を出す日。",
            "呼び出した側が、今度は留守番をする。",
        ],
    },
    "八岐大蛇": {
        "summaryJa": "八つの頭と八つの尾を持つ巨大な蛇。出雲で須佐之男に退治され、尾から草薙の剣が出た。",
        "legendsJa": [
            "年ごとに娘を贄として求める。",
            "八つの酒槽を飲んで眠り、斬られる。",
            "尾の中から剣が出たことが、王家のしるしと結びつく。",
        ],
        "traitsJa": ["八つの頭", "川", "贄", "剣"],
        "creativeHooksJa": [
            "ダムの底に八つの瀬がある。",
            "酒の醸造が封印の一部である村。",
            "剣ではなく、別のものが尾から出る。",
        ],
    },
    "怨霊": {
        "summaryJa": "非業の死を遂げた者の霊が、疫病や天変でこの世に報復する。御霊として祀ると鎮まる、という政治と信仰の語りでもある。",
        "legendsJa": [
            "祀られることで、祟り神から守り神へ変わる。",
            "菅原道真や平将門のように、個別の名を持つ怨霊が歴史を動かす。",
            "無名の怨霊は、疫病の季節にだけ統計へ現れる。",
        ],
        "traitsJa": ["祟り", "御霊", "疫病", "祀り"],
        "creativeHooksJa": [
            "慰霊の手続きが電子化されない理由。",
            "怨霊が和解の条件を文書にする。",
            "祀られない霊の待機列。",
        ],
    },
    "お岩": {
        "summaryJa": "四谷怪談の女。毒で顔を損ない、夫の裏切りを恨んで井戸と家に憑く。歌舞伎以後、日本の怨霊の顔になった。",
        "legendsJa": [
            "産後の体に毒が盛られ、鏡の前で髪が抜ける。",
            "井戸に身を投げたのち、家の夜に顔が浮かぶ。",
            "伊右衛門の行く先に、お岩の気配が先行する。",
        ],
        "traitsJa": ["損われた顔", "井戸", "呪い", "四谷"],
        "creativeHooksJa": [
            "井戸の水質検査が物語を証明する。",
            "お岩が舞台の外で休む日。",
            "顔ではなく声だけが残る改作。",
        ],
    },
    "お菊": {
        "summaryJa": "皿を一枚欠いた責めを受け、井戸に投げられた女中。夜、皿を数える声が井戸から聞こえる番町皿屋敷の霊。",
        "legendsJa": [
            "九枚まで数えて、十枚目で泣く。",
            "欠いた皿は、主人の陰謀だったとする語りがある。",
            "井戸を埋めても、雨の夜に勘定が始まる。",
        ],
        "traitsJa": ["皿", "井戸", "数え声", "女中"],
        "creativeHooksJa": [
            "在庫管理システムが九で止まる店。",
            "十枚目の皿を返す修復家。",
            "お菊が数える対象を皿以外にする。",
        ],
    },
    "なまはげ": {
        "summaryJa": "秋田の来訪神的な鬼面。大晦日などに家を訪れ、怠け者を戒め、子どもを泣かし、年の境を清める。",
        "legendsJa": [
            "「泣く子はいねが」と叫び、家長が豆と酒で迎える。",
            "鬼でありながら、家の福を運ぶ客人でもある。",
            "面と藁の装束は、村ごとに型が違う。",
        ],
        "traitsJa": ["鬼面", "来訪", "戒め", "秋田"],
        "creativeHooksJa": [
            "なまはげが遠隔で勤怠を見る。",
            "泣かない子どもに困った鬼。",
            "面を脱いだあとの静かな茶。",
        ],
    },
    "分福茶釜": {
        "summaryJa": "茂林寺の茶釜に化けた狸の話。茶を沸かすほど福が分かれ、狸の姿に戻っては人を楽しませる。",
        "legendsJa": [
            "僧の茶釜が突然、狸の尾と耳を出す。",
            "見世物になった狸は、芸のあと寺へ帰る。",
            "分けた福は、湯気のように残らないともいう。",
        ],
        "traitsJa": ["茶釜", "化け狸", "福", "湯気"],
        "creativeHooksJa": [
            "無限に沸くわけではない茶の契約。",
            "茶釜が労働基準を申し立てる。",
            "福の分配アルゴリズム。",
        ],
    },
    "龍": {
        "summaryJa": "水を司る巨大な霊獣。雨乞いの対象でも、人を飲み込む淵の主でもある。竜神として祀られる一方、妖怪図鑑では異形の王としても扱われる。",
        "legendsJa": [
            "淵の主が雨を遅らせ、生贄の話になる。",
            "龍が天へ昇るとき、雷と滝が同時に鳴る。",
            "人の姿で里に降り、借りを返す竜神譚もある。",
        ],
        "traitsJa": ["角", "雨", "淵", "昇天"],
        "creativeHooksJa": [
            "水利組合の顧問が龍。",
            "昇天の許可が下りない龍。",
            "雨を一ミリだけ間違える。",
        ],
    },
    "百鬼夜行": {
        "summaryJa": "夜、都の通りを妖怪の列が進む現象。遭遇した者は、隠れるか、護符を持たねば巻き込まれる。",
        "legendsJa": [
            "丑三つ時、楽器と火と道具の霊が行列する。",
            "戸を閉ざし、覗かなかった家だけが無事という。",
            "絵巻は、その列を見世物と警告の両方にした。",
        ],
        "traitsJa": ["行列", "夜", "都", "多数"],
        "creativeHooksJa": [
            "夜行バスが百鬼と車線を共有する。",
            "行列の最後尾が毎回欠ける。",
            "通行止めの理由が百鬼。",
        ],
    },
    "菅原道真": {
        "summaryJa": "藤原氏に左遷され、死後に都で雷と疫を起こしたとされる怨霊。天満の神として祀られ、学問の神へ転じた。",
        "legendsJa": [
            "右近の馬場の梅が、飛んで太宰府へ着いたという飛梅の話。",
            "清涼殿に落雷し、政敵が死んだと記録される。",
            "怒りの神を学問の神へ読み替えたのが、のちの天神信仰である。",
        ],
        "traitsJa": ["雷", "梅", "学問", "御霊"],
        "creativeHooksJa": [
            "受験と落雷が同じ祈願箱。",
            "道真がまだ怒っている年の雷。",
            "左遷の辞令が現代のメールで届く。",
        ],
    },
    "ツチノコ": {
        "summaryJa": "太い胴と短い尾を持つ、蛇に似た未確認の生物。目撃譚と懸賞金で、現代の妖怪とUMAの境にいる。",
        "legendsJa": [
            "山道で、ビール瓶のような胴の蛇を見たという。",
            "跳ねて逃げ、酒の匂いを好むともいう。",
            "村の懸賞金は、観光と真偽のあいだで毎年更新される。",
        ],
        "traitsJa": ["太い胴", "短い尾", "跳躍", "目撃"],
        "creativeHooksJa": [
            "懸賞金の為替だけが動く村。",
            "ツチノコが保護生物になった日。",
            "目撃写真のピントが毎回胴だけ。",
        ],
    },
    "物の怪": {
        "summaryJa": "源氏物語などで、人に憑いて病を起こす見えない気配。特定の顔を持たず、生霊や怨霊の総称としても使われる。",
        "legendsJa": [
            "病床の帳の向こうで、誰かの声がする。",
            "祈祷で正体を移し、別の器へ封じた話がある。",
            "名前を付けると、物の怪は個別の妖怪へ分かれていく。",
        ],
        "traitsJa": ["不可視", "憑依", "病", "総称"],
        "creativeHooksJa": [
            "診断名が物の怪のままのカルテ。",
            "名前を付けるな、と禁じる家。",
            "帳の模様だけが動く病室。",
        ],
    },
}


def is_artifact(s: str) -> bool:
    text = (s or "").strip()
    if not text:
        return False
    lower = text.lower()
    for needle in NEEDLES:
        if needle in lower or needle in text:
            return True
    if lower.startswith("keep ") and " as " in lower:
        return True
    return False


def any_artifact(values) -> bool:
    if values is None:
        return False
    if isinstance(values, str):
        return is_artifact(values)
    if isinstance(values, dict):
        return any(any_artifact(v) for v in values.values())
    return any(any_artifact(v) for v in values)


def clean_list(values, generated, minimum, maximum=None):
    kept = [v for v in (values or []) if isinstance(v, str) and v.strip() and not is_artifact(v)]
    out = []
    seen = set()
    for item in kept + list(generated or []):
        item = item.strip()
        if not item or item in seen or is_artifact(item):
            continue
        seen.add(item)
        out.append(item)
        if maximum and len(out) >= maximum and len(out) >= minimum:
            break
    i = 0
    while len(out) < minimum and generated:
        candidate = generated[i % len(generated)].strip()
        i += 1
        if candidate and candidate not in seen:
            seen.add(candidate)
            out.append(candidate)
        elif i > minimum * 3:
            break
    if maximum:
        return out[:maximum]
    return out


def ordered(obj: dict, keys: list[str]) -> dict:
    out = {}
    for key in keys:
        if key in obj and obj[key] not in (None, "", []):
            out[key] = obj[key]
    for key, value in obj.items():
        if key not in out:
            out[key] = value
    return out


def weak_region(region: str, native: str) -> bool:
    if is_artifact(region):
        return True
    text = region.strip()
    if native and native in text and len(text) <= 10:
        return True
    if any(h in text for h in PLACE_HINTS):
        return False
    return len(text) <= 8 or "巨大" in text or text.endswith("の怪")


def dump(path: Path, value) -> None:
    path.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def main() -> None:
    index = json.loads((DATA / "index.json").read_text())
    profiles = json.loads((DATA / "profiles.json").read_text())
    aliases = json.loads((DATA / "aliases.json").read_text())

    by_native = {e["nativeName"]: e for e in index}

    for entry in index:
        native = entry["nativeName"]
        if is_artifact(entry.get("blurbJa", "")):
            cat = entry.get("category", "妖怪")
            tags = [t for t in (entry.get("tags") or []) if t]
            hint = "、".join(tags[:3])
            if hint:
                entry["blurbJa"] = f"{native}。{hint}にまつわる{cat}の伝承。"
            else:
                entry["blurbJa"] = f"{native}。{cat}として語られる存在。"
        if weak_region(entry.get("region", ""), native):
            entry["region"] = CAT_REGION.get(entry.get("category", ""), "日本")
    new_index = [ordered(e, INDEX_ORDER) for e in index]

    for profile in profiles:
        native = profile["nativeName"]
        entry = by_native.get(native, {})
        blurb = (entry.get("blurbJa") or profile.get("summaryJa") or native).rstrip("。")
        category = profile.get("category") or entry.get("category") or "妖怪"
        cat_en = profile.get("categoryEn") or CAT_EN.get(category, "yokai")
        tags = entry.get("tags") or []
        name = profile.get("name") or native
        dirty = any_artifact(profile)

        generated_legends_ja = [
            f"{blurb}。",
            f"{native}は{category}に数えられ、土地の話に現れる。",
            f"名を丁寧に呼び、刺激しないのが無難とされる。",
        ]
        generated_legends = [
            f"{name} appears in Japanese {cat_en} lore.",
            blurb + ".",
            f"Courtesy is safer than curiosity when {name} is near.",
        ]
        generated_traits = [t for t in tags[:4] if t] + [category, cat_en]
        generated_motifs = [category] + [t for t in tags[:3] if t] + ["encounter"]
        generated_hooks = [
            f"Write a present-day scene that still obeys the old rule about {native}.",
            f"A mapmaker realizes the terrain is shaped like {native}.",
            f"A festival, kitchen, or night road where {native} is the unwritten rule.",
        ]

        if dirty or is_artifact(profile.get("summaryJa", "")) or len(profile.get("summaryJa", "")) < 16:
            extra = f"{category}の伝承に名が残る。"
            profile["summaryJa"] = blurb if blurb.endswith("。") else blurb + "。"
            if extra not in profile["summaryJa"]:
                profile["summaryJa"] += extra
        if dirty or is_artifact(profile.get("summary", "")):
            profile["summary"] = f"{name} is a {cat_en} in Japanese folklore. {blurb}."
        if dirty or is_artifact(profile.get("funFactJa", "")) or len(profile.get("funFactJa", "")) < 6:
            profile["funFactJa"] = f"{native}は、短い注意の言葉として語り継がれる。"
        if dirty or is_artifact(profile.get("funFact", "")):
            profile["funFact"] = f"The shortest warning about {name} is often the one that survives."
        if dirty or any_artifact(profile.get("legends")):
            profile["legends"] = clean_list(profile.get("legends"), generated_legends, 3, 4)
        if dirty or any_artifact(profile.get("traits")):
            profile["traits"] = clean_list(profile.get("traits"), generated_traits, 4, 5)
        if dirty or any_artifact(profile.get("motifs")):
            profile["motifs"] = clean_list(profile.get("motifs"), generated_motifs, 4, 5)
        if dirty or any_artifact(profile.get("creativeHooks")):
            profile["creativeHooks"] = clean_list(profile.get("creativeHooks"), generated_hooks, 3, 3)
        if dirty or any_artifact(profile.get("sources")) or not profile.get("sources"):
            profile["sources"] = [
                s for s in (profile.get("sources") or []) if not is_artifact(s)
            ] or ["Folklore index card; primary citation not yet attached"]
        if weak_region(profile.get("region", ""), native):
            profile["region"] = CAT_REGION.get(category, profile.get("region", "Japan"))

        overlay = RANK1.get(native)
        if overlay:
            profile.update({k: v for k, v in overlay.items() if v})
        elif dirty:
            profile["legendsJa"] = clean_list(profile.get("legendsJa"), generated_legends_ja, 3, 3)

        if entry.get("famousRank") == 1 and len(profile.get("summaryJa", "")) < 40:
            profile["summaryJa"] = f"{blurb}。{category}を代表する存在として、いまも名前が通る。"

    new_profiles = [ordered(p, PROFILE_ORDER) for p in profiles]

    aliases["狐"] = "九尾の狐"
    aliases["きつね"] = "九尾の狐"
    aliases = dict(sorted(aliases.items(), key=lambda kv: kv[0]))

    dump(DATA / "index.json", new_index)
    dump(DATA / "profiles.json", new_profiles)
    dump(DATA / "aliases.json", aliases)

    residue = 0
    for p in new_profiles:
        blob = json.dumps(p, ensure_ascii=False)
        if is_artifact(blob):
            residue += 1
            print("still dirty:", p["nativeName"])
    print(f"wrote {len(new_index)} index, {len(new_profiles)} profiles, residue={residue}")


if __name__ == "__main__":
    main()
