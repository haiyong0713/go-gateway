package search

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-intl/interface/conf"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

const (
	_searchJSON = `{
		"seid": "13552721685685175391",
		"page": 1,
		"pagesize": 20,
		"numResults": 9,
		"numPages": 1,
		"exp_bits": 1,
		"pageinfo": {
			"upuser": {
				"numResults": 0,
				"pages": 0
			},
			"bili_user": {
				"numResults": 0,
				"pages": 0
			},
			"user": {
				"numResults": 0,
				"pages": 0
			},
			"movie": {
				"numResults": 0,
				"pages": 0
			},
			"article": {
				"numResults": 0,
				"pages": 0
			},
			"live_room": {
				"numResults": 0,
				"pages": 0
			},
			"live_user": {
				"numResults": 0,
				"pages": 0
			},
			"live_all": {
				"numResults": 0,
				"pages": 0
			},
			"media_bangumi": {
				"numResults": 0,
				"pages": 0
			},
			"media_ft": {
				"numResults": 0,
				"pages": 0
			}
		},
		"result": {
			"video": [
				{
					"id": 67024960,
					"author": "火线啊啊啊",
					"title": "【\u003cem class=\"keyword\"\u003esj\u003c/em\u003e \u003cem class=\"keyword\"\u003ereturns\u003c/em\u003e3】预告\u003cem class=\"keyword\"\u003e4\u003c/em\u003e",
					"pic": "https://i0.hdslb.com/bfs/archive/e68cb05d02a0b32d8f1a8a3c6472e96fa379326e.jpg",
					"description": "Vapp\nSuju冲压，三连一个哦",
					"play": 6254,
					"video_review": 1,
					"duration": "0:58",
					"numPages": 0,
					"view_type": "",
					"rec_tags": null,
					"new_rec_tags": [
						
					]
				},
				{
					"id": 15514687,
					"author": "나비야",
					"title": "《模拟人生\u003cem class=\"keyword\"\u003e4\u003c/em\u003e速建》警告！\u003cem class=\"keyword\"\u003esuperjunior\u003c/em\u003e粉丝的房间竟是这样",
					"pic": "https://i0.hdslb.com/bfs/archive/5842c429c907fd9b467a1b25783f41d608f7fa83.jpg",
					"description": "我们老少年终于要回归了,太激动惹！\n给老少年打钱先满足下自己哈哈哈，dreaming a house like that\nPS:大家别忘了买专辑啊，SJ returns也要刷起来~别忘了用V app看哦~",
					"play": 4268,
					"video_review": 124,
					"duration": "18:19",
					"numPages": 0,
					"view_type": "",
					"rec_tags": null,
					"new_rec_tags": [
						
					]
				},
				{
					"id": 67020442,
					"author": "Yeon0629",
					"title": "\u003cem class=\"keyword\"\u003eSuperjunior\u003c/em\u003e新综艺《\u003cem class=\"keyword\"\u003eSJ\u003c/em\u003e \u003cem class=\"keyword\"\u003eRETURNS\u003c/em\u003e 3》预告\u003cem class=\"keyword\"\u003e4\u003c/em\u003e公开～",
					"pic": "https://i0.hdslb.com/bfs/archive/c8c4212012f00e0ba3f707ec5d006289595c32fd.jpg",
					"description": "Cr.Vlive",
					"play": 2348,
					"video_review": 0,
					"duration": "0:59",
					"numPages": 0,
					"view_type": "",
					"rec_tags": null,
					"new_rec_tags": [
						
					]
				},
				{
					"id": 67038516,
					"author": "随便字幕",
					"title": "\u003cem class=\"keyword\"\u003eSJ\u003c/em\u003e \u003cem class=\"keyword\"\u003eReturns\u003c/em\u003e3 预告\u003cem class=\"keyword\"\u003e4\u003c/em\u003e SUJU也期待的~ 明日11点就要和大家见面！",
					"pic": "https://i0.hdslb.com/bfs/archive/1e289e35643326a31aaf2be3953156a267855daa.jpg",
					"description": "vlive\n明日中午韩国时间12点，北京时间11点Vlive更新哦~",
					"play": 4124,
					"video_review": 0,
					"duration": "0:59",
					"numPages": 0,
					"view_type": "",
					"rec_tags": null,
					"new_rec_tags": [
						
					]
				},
				{
					"id": 20632953,
					"author": "唯愛利特",
					"title": "Super Junior-突襲放送吃貨\u003cem class=\"keyword\"\u003e4\u003c/em\u003e人組",
					"pic": "https://i0.hdslb.com/bfs/archive/aa207443fe8dd031551bd00fe733d9c1192a8b8a.jpg",
					"description": "慶祝SJ returns NAVER 點擊率破2300萬",
					"play": 23775,
					"video_review": 753,
					"duration": "23:18",
					"numPages": 0,
					"view_type": "",
					"rec_tags": null,
					"new_rec_tags": [
						
					]
				},
				{
					"id": 35660352,
					"author": "我萌的cp都BE了",
					"title": "木浦小可爱\u003cem class=\"keyword\"\u003e4\u003c/em\u003e",
					"pic": "https://i0.hdslb.com/bfs/archive/avsas_i181110tx1lw98do26yxyrdg8acifiad_0023.jpg",
					"description": "SJ Returns 下篇",
					"play": 5967,
					"video_review": 30,
					"duration": "17:0",
					"numPages": 0,
					"view_type": "",
					"rec_tags": null,
					"new_rec_tags": [
						
					]
				},
				{
					"id": 19247099,
					"author": "kiwiii-",
					"title": "李东嘿：我再也不是你们的宝宝了",
					"pic": "https://i0.hdslb.com/bfs/archive/12f745467eb27a3ec70fb660c040c163cc66648e.png",
					"description": "cr: ytb\r\np1~4全员模仿东海，p4完整版戳av18814796的p4\r\np5: super tv cut\r\np6: sj returns cut\r\np7: 生日被闹",
					"play": 32390,
					"video_review": 333,
					"duration": "27:50",
					"numPages": 0,
					"view_type": "",
					"rec_tags": null,
					"new_rec_tags": [
						
					]
				},
				{
					"id": 40438627,
					"author": "雯雯雯崽",
					"title": "【澈海】希澈：我最喜欢这样的弟弟啦，又可爱又会撒娇",
					"pic": "https://i0.hdslb.com/bfs/archive/6c3757c79135b0ca00a3201dc639adce32641ea0.jpg",
					"description": "1.电视购物卖面膜 百蓝能找到\n2.a song for you\n3.sj returns\n4.一周的偶像\n好像不是按顺序的\nbgm：恋爱的开始(韩语）",
					"play": 39745,
					"video_review": 89,
					"duration": "3:12",
					"numPages": 0,
					"view_type": "",
					"rec_tags": null,
					"new_rec_tags": [
						
					]
				},
				{
					"id": 57000144,
					"author": "_三十七",
					"title": "【沙雕预警】就没有你司机踩不上的点（综艺混剪｜沙雕踩点）",
					"pic": "https://i0.hdslb.com/bfs/archive/2761ba525c9a12a47b6a6277baaf7dd05c9bad12.jpg",
					"description": "纯属娱乐，不喜轻喷\n如有不妥，请多指正\n\n素材源自字幕组：\n SUPER TV 第一季、第二季\n SJ Returns\n SNL\n Super Camp\n 新西游记\n Super Show4\n 认识的哥哥",
					"play": 50791,
					"video_review": 114,
					"duration": "1:31",
					"numPages": 0,
					"view_type": "",
					"rec_tags": [
						"播放量较多"
					],
					"new_rec_tags": [
						{
							"tag_name": "播放量较多",
							"tag_style": 1
						}
					]
				}
			]
		},
		"flow_result": [
			{
				"linktype": "video",
				"position": 1,
				"type": "video",
				"value": {
					"rank_offset": 1,
					"play": 6254,
					"new_rec_tags": [
						
					],
					"description": "Vapp\nSuju\u51b2\u538b\uff0c\u4e09\u8fde\u4e00\u4e2a\u54e6",
					"pubdate": 1567915801,
					"view_type": "",
					"arcrank": "0",
					"pic": "https:\/\/i0.hdslb.com\/bfs\/archive\/e68cb05d02a0b32d8f1a8a3c6472e96fa379326e.jpg",
					"tag": "SUJU,SJ,super junior,KOREA\u76f8\u5173,\u9884\u544a,sj returns",
					"video_review": 1,
					"is_pay": 0,
					"favorites": 10,
					"rank_index": 1,
					"duration": "0:58",
					"id": 67024960,
					"rank_score": 100459,
					"badgepay": false,
					"typeid": "131",
					"senddate": 1567915802,
					"title": "\u3010\u003cem class=\"keyword\"\u003esj\u003c\/em\u003e \u003cem class=\"keyword\"\u003ereturns\u003c\/em\u003e3\u3011\u9884\u544a\u003cem class=\"keyword\"\u003e4\u003c\/em\u003e",
					"review": 3,
					"author": "\u706b\u7ebf\u554a\u554a\u554a",
					"hit_columns": [
						"title",
						"tag"
					],
					"mid": 276167511,
					"is_union_video": 0,
					"arcurl": "http:\/\/www.bilibili.com\/video\/av67024960",
					"typename": "Korea\u76f8\u5173",
					"aid": 67024960,
					"type": "video",
					"rec_tags": null
				},
				"Video": null,
				"Live": null,
				"Operate": null,
				"Article": null,
				"Media": null,
				"User": null,
				"Game": null,
				"Query": null,
				"Twitter": null,
				"trackid": "13552721685685175391"
			},
			{
				"linktype": "video",
				"position": 2,
				"type": "video",
				"value": {
					"rank_offset": 2,
					"play": 4268,
					"new_rec_tags": [
						
					],
					"description": "\u6211\u4eec\u8001\u5c11\u5e74\u7ec8\u4e8e\u8981\u56de\u5f52\u4e86,\u592a\u6fc0\u52a8\u60f9\uff01\n\u7ed9\u8001\u5c11\u5e74\u6253\u94b1\u5148\u6ee1\u8db3\u4e0b\u81ea\u5df1\u54c8\u54c8\u54c8\uff0cdreaming a house like that\nPS:\u5927\u5bb6\u522b\u5fd8\u4e86\u4e70\u4e13\u8f91\u554a\uff0cSJ returns\u4e5f\u8981\u5237\u8d77\u6765~\u522b\u5fd8\u4e86\u7528V app\u770b\u54e6~",
					"pubdate": 1508333572,
					"view_type": "",
					"arcrank": "0",
					"pic": "https:\/\/i0.hdslb.com\/bfs\/archive\/5842c429c907fd9b467a1b25783f41d608f7fa83.jpg",
					"tag": "ELF,1106\u56de\u5f52,\u5355\u673a\u6e38\u620f,\u6a21\u62df\u4eba\u751f4,superjunior",
					"video_review": 124,
					"is_pay": 0,
					"favorites": 136,
					"rank_index": 2,
					"duration": "18:19",
					"id": 15514687,
					"rank_score": 100426,
					"badgepay": false,
					"typeid": "17",
					"senddate": 1543770125,
					"title": "\u300a\u6a21\u62df\u4eba\u751f\u003cem class=\"keyword\"\u003e4\u003c\/em\u003e\u901f\u5efa\u300b\u8b66\u544a\uff01\u003cem class=\"keyword\"\u003esuperjunior\u003c\/em\u003e\u7c89\u4e1d\u7684\u623f\u95f4\u7adf\u662f\u8fd9\u6837",
					"review": 100,
					"author": "\ub098\ube44\uc57c",
					"hit_columns": [
						"title",
						"description",
						"tag"
					],
					"mid": 62105397,
					"is_union_video": 0,
					"arcurl": "http:\/\/www.bilibili.com\/video\/av15514687",
					"typename": "\u5355\u673a\u6e38\u620f",
					"aid": 15514687,
					"type": "video",
					"rec_tags": null
				},
				"Video": null,
				"Live": null,
				"Operate": null,
				"Article": null,
				"Media": null,
				"User": null,
				"Game": null,
				"Query": null,
				"Twitter": null,
				"trackid": "13552721685685175391"
			},
			{
				"linktype": "video",
				"position": 3,
				"type": "video",
				"value": {
					"rank_offset": 3,
					"play": 2348,
					"new_rec_tags": [
						
					],
					"description": "Cr.Vlive",
					"pubdate": 1567913213,
					"view_type": "",
					"arcrank": "0",
					"pic": "https:\/\/i0.hdslb.com\/bfs\/archive\/c8c4212012f00e0ba3f707ec5d006289595c32fd.jpg",
					"tag": "SUPERJUNIOR,SJ,KOREA\u76f8\u5173,sj return3,superjunior,SJ return",
					"video_review": 0,
					"is_pay": 0,
					"favorites": 3,
					"rank_index": 3,
					"duration": "0:59",
					"id": 67020442,
					"rank_score": 100263,
					"badgepay": false,
					"typeid": "71",
					"senddate": 1567964700,
					"title": "\u003cem class=\"keyword\"\u003eSuperjunior\u003c\/em\u003e\u65b0\u7efc\u827a\u300a\u003cem class=\"keyword\"\u003eSJ\u003c\/em\u003e \u003cem class=\"keyword\"\u003eRETURNS\u003c\/em\u003e 3\u300b\u9884\u544a\u003cem class=\"keyword\"\u003e4\u003c\/em\u003e\u516c\u5f00\uff5e",
					"review": 1,
					"author": "Yeon0629",
					"hit_columns": [
						"title",
						"tag"
					],
					"mid": 385258465,
					"is_union_video": 0,
					"arcurl": "http:\/\/www.bilibili.com\/video\/av67020442",
					"typename": "\u7efc\u827a",
					"aid": 67020442,
					"type": "video",
					"rec_tags": null
				},
				"Video": null,
				"Live": null,
				"Operate": null,
				"Article": null,
				"Media": null,
				"User": null,
				"Game": null,
				"Query": null,
				"Twitter": null,
				"trackid": "13552721685685175391"
			},
			{
				"linktype": "video",
				"position": 4,
				"type": "video",
				"value": {
					"rank_offset": 4,
					"play": 4124,
					"new_rec_tags": [
						
					],
					"description": "vlive\n\u660e\u65e5\u4e2d\u5348\u97e9\u56fd\u65f6\u95f412\u70b9\uff0c\u5317\u4eac\u65f6\u95f411\u70b9Vlive\u66f4\u65b0\u54e6~",
					"pubdate": 1567923600,
					"view_type": "",
					"arcrank": "0",
					"pic": "https:\/\/i0.hdslb.com\/bfs\/archive\/1e289e35643326a31aaf2be3953156a267855daa.jpg",
					"tag": "SJ,SJ Returns3,Super Junior,SUJU,\u5e0c\u6f88,\u827a\u58f0,\u94f6\u8d6b,\u795e\u7ae5,\u572d\u8d24,\u4e1c\u6d77",
					"video_review": 0,
					"is_pay": 0,
					"favorites": 7,
					"rank_index": 4,
					"duration": "0:59",
					"id": 67038516,
					"rank_score": 100030,
					"badgepay": false,
					"typeid": "131",
					"senddate": 1567927625,
					"title": "\u003cem class=\"keyword\"\u003eSJ\u003c\/em\u003e \u003cem class=\"keyword\"\u003eReturns\u003c\/em\u003e3 \u9884\u544a\u003cem class=\"keyword\"\u003e4\u003c\/em\u003e SUJU\u4e5f\u671f\u5f85\u7684~ \u660e\u65e511\u70b9\u5c31\u8981\u548c\u5927\u5bb6\u89c1\u9762\uff01",
					"review": 2,
					"author": "\u968f\u4fbf\u5b57\u5e55",
					"hit_columns": [
						"title",
						"tag"
					],
					"mid": 16280611,
					"is_union_video": 0,
					"arcurl": "http:\/\/www.bilibili.com\/video\/av67038516",
					"typename": "Korea\u76f8\u5173",
					"aid": 67038516,
					"type": "video",
					"rec_tags": null
				},
				"Video": null,
				"Live": null,
				"Operate": null,
				"Article": null,
				"Media": null,
				"User": null,
				"Game": null,
				"Query": null,
				"Twitter": null,
				"trackid": "13552721685685175391"
			},
			{
				"linktype": "video",
				"position": 5,
				"type": "video",
				"value": {
					"rank_offset": 5,
					"play": 23775,
					"new_rec_tags": [
						
					],
					"description": "\u6176\u795dSJ returns NAVER \u9ede\u64ca\u7387\u78342300\u842c",
					"pubdate": 1520748149,
					"view_type": "",
					"arcrank": "0",
					"pic": "https:\/\/i0.hdslb.com\/bfs\/archive\/aa207443fe8dd031551bd00fe733d9c1192a8b8a.jpg",
					"tag": "SJ,Korea\u76f8\u5173,\u85dd\u8072,\u9280\u8d6b,\u795e\u7ae5,\u6771\u6d77",
					"video_review": 753,
					"is_pay": 0,
					"favorites": 2368,
					"rank_index": 5,
					"duration": "23:18",
					"id": 20632953,
					"rank_score": 101803,
					"badgepay": false,
					"typeid": "131",
					"senddate": 1520748149,
					"title": "Super Junior-\u7a81\u8972\u653e\u9001\u5403\u8ca8\u003cem class=\"keyword\"\u003e4\u003c\/em\u003e\u4eba\u7d44",
					"review": 41,
					"author": "\u552f\u611b\u5229\u7279",
					"hit_columns": [
						"title",
						"description",
						"tag"
					],
					"mid": 283038075,
					"is_union_video": 0,
					"arcurl": "http:\/\/www.bilibili.com\/video\/av20632953",
					"typename": "Korea\u76f8\u5173",
					"aid": 20632953,
					"type": "video",
					"rec_tags": null
				},
				"Video": null,
				"Live": null,
				"Operate": null,
				"Article": null,
				"Media": null,
				"User": null,
				"Game": null,
				"Query": null,
				"Twitter": null,
				"trackid": "13552721685685175391"
			},
			{
				"linktype": "video",
				"position": 6,
				"type": "video",
				"value": {
					"rank_offset": 6,
					"play": 5967,
					"new_rec_tags": [
						
					],
					"description": "SJ Returns \u4e0b\u7bc7",
					"pubdate": 1541822817,
					"view_type": "",
					"arcrank": "0",
					"pic": "https:\/\/i0.hdslb.com\/bfs\/archive\/avsas_i181110tx1lw98do26yxyrdg8acifiad_0023.jpg",
					"tag": "Korea\u76f8\u5173,\u674e\u4e1c\u6d77,superjunior,sj ruturns",
					"video_review": 30,
					"is_pay": 0,
					"favorites": 162,
					"rank_index": 6,
					"duration": "17:0",
					"id": 35660352,
					"rank_score": 101052,
					"badgepay": false,
					"typeid": "131",
					"senddate": 1556963433,
					"title": "\u6728\u6d66\u5c0f\u53ef\u7231\u003cem class=\"keyword\"\u003e4\u003c\/em\u003e",
					"review": 7,
					"author": "\u6211\u840c\u7684cp\u90fdBE\u4e86",
					"hit_columns": [
						"title",
						"description",
						"tag"
					],
					"mid": 100690776,
					"is_union_video": 0,
					"arcurl": "http:\/\/www.bilibili.com\/video\/av35660352",
					"typename": "Korea\u76f8\u5173",
					"aid": 35660352,
					"type": "video",
					"rec_tags": null
				},
				"Video": null,
				"Live": null,
				"Operate": null,
				"Article": null,
				"Media": null,
				"User": null,
				"Game": null,
				"Query": null,
				"Twitter": null,
				"trackid": "13552721685685175391"
			},
			{
				"linktype": "video",
				"position": 7,
				"type": "video",
				"value": {
					"rank_offset": 7,
					"play": 32390,
					"new_rec_tags": [
						
					],
					"description": "cr: ytb\r\np1~4\u5168\u5458\u6a21\u4eff\u4e1c\u6d77\uff0cp4\u5b8c\u6574\u7248\u6233av18814796\u7684p4\r\np5: super tv cut\r\np6: sj returns cut\r\np7: \u751f\u65e5\u88ab\u95f9",
					"pubdate": 1517956482,
					"view_type": "",
					"arcrank": "0",
					"pic": "https:\/\/i0.hdslb.com\/bfs\/archive\/12f745467eb27a3ec70fb660c040c163cc66648e.png",
					"tag": "superjunior,\u674e\u4e1c\u6d77,\u79fb\u52a8\u563f,\u8001\u5c11\u5e74,suju",
					"video_review": 333,
					"is_pay": 0,
					"favorites": 1469,
					"rank_index": 7,
					"duration": "27:50",
					"id": 19247099,
					"rank_score": 100573,
					"badgepay": false,
					"typeid": "131",
					"senddate": 1532591481,
					"title": "\u674e\u4e1c\u563f\uff1a\u6211\u518d\u4e5f\u4e0d\u662f\u4f60\u4eec\u7684\u5b9d\u5b9d\u4e86",
					"review": 28,
					"author": "kiwiii-",
					"hit_columns": [
						"description",
						"tag"
					],
					"mid": 956763,
					"is_union_video": 0,
					"arcurl": "http:\/\/www.bilibili.com\/video\/av19247099",
					"typename": "Korea\u76f8\u5173",
					"aid": 19247099,
					"type": "video",
					"rec_tags": null
				},
				"Video": null,
				"Live": null,
				"Operate": null,
				"Article": null,
				"Media": null,
				"User": null,
				"Game": null,
				"Query": null,
				"Twitter": null,
				"trackid": "13552721685685175391"
			},
			{
				"linktype": "video",
				"position": 8,
				"type": "video",
				"value": {
					"rank_offset": 8,
					"play": 39745,
					"new_rec_tags": [
						
					],
					"description": "1.\u7535\u89c6\u8d2d\u7269\u5356\u9762\u819c \u767e\u84dd\u80fd\u627e\u5230\n2.a song for you\n3.sj returns\n4.\u4e00\u5468\u7684\u5076\u50cf\n\u597d\u50cf\u4e0d\u662f\u6309\u987a\u5e8f\u7684\nbgm\uff1a\u604b\u7231\u7684\u5f00\u59cb(\u97e9\u8bed\uff09",
					"pubdate": 1547174563,
					"view_type": "",
					"arcrank": "0",
					"pic": "https:\/\/i0.hdslb.com\/bfs\/archive\/6c3757c79135b0ca00a3201dc639adce32641ea0.jpg",
					"tag": "superjunior,\u91d1\u5e0c\u6f88,\u674e\u4e1c\u6d77,\u674e\u8d6b\u5bb0,\u5229\u7279,\u66fa\u572d\u8d24",
					"video_review": 89,
					"is_pay": 0,
					"favorites": 873,
					"rank_index": 8,
					"duration": "3:12",
					"id": 40438627,
					"rank_score": 100462,
					"badgepay": false,
					"typeid": "131",
					"senddate": 1547459739,
					"title": "\u3010\u6f88\u6d77\u3011\u5e0c\u6f88\uff1a\u6211\u6700\u559c\u6b22\u8fd9\u6837\u7684\u5f1f\u5f1f\u5566\uff0c\u53c8\u53ef\u7231\u53c8\u4f1a\u6492\u5a07",
					"review": 36,
					"author": "\u96ef\u96ef\u96ef\u5d3d",
					"hit_columns": [
						"description",
						"tag"
					],
					"mid": 31966247,
					"is_union_video": 0,
					"arcurl": "http:\/\/www.bilibili.com\/video\/av40438627",
					"typename": "Korea\u76f8\u5173",
					"aid": 40438627,
					"type": "video",
					"rec_tags": null
				},
				"Video": null,
				"Live": null,
				"Operate": null,
				"Article": null,
				"Media": null,
				"User": null,
				"Game": null,
				"Query": null,
				"Twitter": null,
				"trackid": "13552721685685175391"
			},
			{
				"linktype": "video",
				"position": 9,
				"type": "video",
				"value": {
					"rank_offset": 9,
					"play": 50791,
					"new_rec_tags": [
						{
							"tag_name": "\u64ad\u653e\u91cf\u8f83\u591a",
							"tag_style": 1
						}
					],
					"description": "\u7eaf\u5c5e\u5a31\u4e50\uff0c\u4e0d\u559c\u8f7b\u55b7\n\u5982\u6709\u4e0d\u59a5\uff0c\u8bf7\u591a\u6307\u6b63\n\n\u7d20\u6750\u6e90\u81ea\u5b57\u5e55\u7ec4\uff1a\n SUPER TV \u7b2c\u4e00\u5b63\u3001\u7b2c\u4e8c\u5b63\n SJ Returns\n SNL\n Super Camp\n \u65b0\u897f\u6e38\u8bb0\n Super Show4\n \u8ba4\u8bc6\u7684\u54e5\u54e5",
					"pubdate": 1561620642,
					"view_type": "",
					"arcrank": "0",
					"pic": "https:\/\/i0.hdslb.com\/bfs\/archive\/2761ba525c9a12a47b6a6277baaf7dd05c9bad12.jpg",
					"tag": "\u7efc\u827a,SUPERTV,KOREA\u76f8\u5173,SJ,\u641e\u7b11,SUPERTV2,\u6df7\u526a,\u7efc\u827a\u526a\u8f91,SUPER JUNIOR,Korea\u76f8\u5173",
					"video_review": 114,
					"is_pay": 0,
					"favorites": 1373,
					"rank_index": 9,
					"duration": "1:31",
					"id": 57000144,
					"rank_score": 100228,
					"badgepay": false,
					"typeid": "71",
					"senddate": 1561739461,
					"title": "\u3010\u6c99\u96d5\u9884\u8b66\u3011\u5c31\u6ca1\u6709\u4f60\u53f8\u673a\u8e29\u4e0d\u4e0a\u7684\u70b9\uff08\u7efc\u827a\u6df7\u526a\uff5c\u6c99\u96d5\u8e29\u70b9\uff09",
					"review": 77,
					"author": "_\u4e09\u5341\u4e03",
					"hit_columns": [
						"description",
						"tag"
					],
					"mid": 4202219,
					"is_union_video": 0,
					"arcurl": "http:\/\/www.bilibili.com\/video\/av57000144",
					"typename": "\u7efc\u827a",
					"aid": 57000144,
					"type": "video",
					"rec_tags": [
						"\u64ad\u653e\u91cf\u8f83\u591a"
					]
				},
				"Video": null,
				"Live": null,
				"Operate": null,
				"Article": null,
				"Media": null,
				"User": null,
				"Game": null,
				"Query": null,
				"Twitter": null,
				"trackid": "13552721685685175391"
			},
			{
				"linktype": "query_rec",
				"position": 10,
				"type": "query",
				"type_name": "相关推荐",
				"value": [
					{
						"from_source": "query_rec_search",
						"type": "query_rec",
						"name": "\u5389\u65ed",
						"id": 5257065857987777979
					},
					{
						"from_source": "query_rec_search",
						"type": "query_rec",
						"name": "\u827a\u58f0",
						"id": 2261245828994134783
					},
					{
						"from_source": "query_rec_search",
						"type": "query_rec",
						"name": "SHINHWA",
						"id": 2150928615163505855
					},
					{
						"from_source": "query_rec_search",
						"type": "query_rec",
						"name": "FOODIELOVEY",
						"id": 4209113022151721846
					},
					{
						"from_source": "query_rec_search",
						"type": "query_rec",
						"name": "\u5c71\u72d7",
						"id": 228533137885070116
					},
					{
						"from_source": "query_rec_search",
						"type": "query_rec",
						"name": "\u674e\u4e1c\u6d77",
						"id": 8782829756625984837
					}
				],
				"Video": null,
				"Live": null,
				"Operate": null,
				"Article": null,
				"Media": null,
				"User": null,
				"Game": null,
				"Query": null,
				"Twitter": null,
				"trackid": "13552721685685175391"
			}
		],
		"flow_placeholder": 1
	}`
)

var (
	d *Dao
)

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestMain(m *testing.M) {
	if os.Getenv("DEPLOY_ENV") != "" {
		flag.Set("app_id", "main.app-svr.app-intl")
		flag.Set("conf_token", "02007e8d0f77d31baee89acb5ce6d3ac")
		flag.Set("tree_id", "64518")
		flag.Set("conf_version", "docker-1")
		flag.Set("deploy_env", "uat")
		flag.Set("conf_host", "config.bilibili.co")
		flag.Set("conf_path", "/tmp")
		flag.Set("region", "sh")
		flag.Set("zone", "sh001")
	} else {
		flag.Set("conf", "../../cmd/app-intl-test.toml")
	}
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
}

func TestSearch(t *testing.T) {
	Convey("get Search", t, func() {
		var (
			mid, zoneid                                                                                                                               int64
			mobiApp, device, platform, buvid, keyword, duration, order, filtered, fromSource, recommend                                               string
			plat                                                                                                                                      int8
			seasonNum, movieNum, upUserNum, uvLimit, userNum, userVideoLimit, biliUserNum, biliUserVideoLimit, rid, highlight, build, pn, ps, isQuery int
			now                                                                                                                                       time.Time
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.main).Reply(200).JSON(_searchJSON)
		res, _, err := d.Search(context.Background(), mid, zoneid, mobiApp, device, platform, buvid, keyword, duration, order, filtered, fromSource, recommend, plat, seasonNum, movieNum, upUserNum, uvLimit, userNum, userVideoLimit, biliUserNum, biliUserVideoLimit, rid, highlight, build, pn, ps, isQuery, now)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestSeason2(t *testing.T) {
	Convey("Season2", t, func() {
		var (
			id                                        int64
			keyword, mobiApp, device, platform, buvid string
			highlight, build, pn, ps                  int
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.main).Reply(200).JSON(_searchJSON)
		res, err := d.Season2(context.Background(), id, keyword, mobiApp, device, platform, buvid, highlight, build, pn, ps)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestMovieByType2(t *testing.T) {
	Convey("MovieByType2", t, func() {
		var (
			mid                                       int64
			keyword, mobiApp, device, platform, buvid string
			highlight, build, pn, ps                  int
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.main).Reply(200).JSON(_searchJSON)
		res, err := d.MovieByType2(context.Background(), mid, keyword, mobiApp, device, platform, buvid, highlight, build, pn, ps)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestUpper(t *testing.T) {
	Convey("Upper", t, func() {
		var (
			mid                                                        int64
			keyword, mobiApp, device, platform, buvid, filtered, order string
			biliUserVL, highlight, build, userType, orderSort, pn, ps  int
			now                                                        time.Time
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.main).Reply(200).JSON(_searchJSON)
		res, err := d.Upper(context.Background(), mid, keyword, mobiApp, device, platform, buvid, filtered, order, biliUserVL, highlight, build, userType, orderSort, pn, ps, now)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestArticleByType(t *testing.T) {
	Convey("ArticleByType", t, func() {
		var (
			mid, zoneid                                                       int64
			keyword, mobiApp, device, platform, buvid, filtered, order, sType string
			plat                                                              int8
			categoryID, build, highlight, pn, ps                              int
			now                                                               time.Time
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.main).Reply(200).JSON(_searchJSON)
		res, err := d.ArticleByType(context.Background(), mid, zoneid, keyword, mobiApp, device, platform, buvid, filtered, order, sType, plat, categoryID, build, highlight, pn, ps, now)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestChannel(t *testing.T) {
	Convey("Channel", t, func() {
		var (
			mid                                                     int64
			keyword, mobiApp, platform, buvid, device, order, sType string
			build, pn, ps, highlight                                int
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.main).Reply(200).JSON(_searchJSON)
		res, err := d.Channel(context.Background(), mid, keyword, mobiApp, platform, buvid, device, order, sType, build, pn, ps, highlight)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

func TestSuggest3(t *testing.T) {
	Convey("Suggest3", t, func() {
		var (
			mid                   int64
			platform, buvid, term string
			build, highlight      int
			mobiApp               string
			now                   time.Time
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.main).Reply(200).JSON(_searchJSON)
		res, err := d.Suggest3(context.Background(), mid, platform, buvid, term, build, highlight, mobiApp, now)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}
