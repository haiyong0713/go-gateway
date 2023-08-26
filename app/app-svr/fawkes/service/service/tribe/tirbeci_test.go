package tribe

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var newJsonStr = `{
  "name": "omg",
  "versionCode": 11,
  "versionName": "1.0",
  "priority": 0,
  "builtIn": false,
  "forceDependencyVersion": true,
  "modules": [],
  "components": [],
  "dataBindingPackages": [],
  "dependencies": [ 
    {
      "name": "host",
      "versionCode": 3,
      "depFeature": "infra"
    }
  ],
  "packageId": 127,
  "features": [ 
    {
      "featureName": "default",
      "compatibleVersions": [
        9,
        10
      ]
    },
    {
      "featureName": "infra",
      "compatibleVersions": []
    }
  ]
}
`

var oldJsonStr = `{
"name": "gametribe",
"versionCode": 735920600,
"versionName": "0.0.0",
"priority": 0,
"builtIn": true,
"forceDependencyVersion": true,
"modules": [
{
"name": "game_tribe",
"entranceClass": "com.bilibili.lib.blrouter.internal.generated.Game_tribe",
"attributes": []
}
],
"components": [
{
"name": "com.bilibili.biligame.GameInformationListActivity",
"process": "",
"type": "activity"
},
{
"name": "com.bilibili.biligame.ui.gamelist.GameBookCenterActivityV2",
"process": "",
"type": "activity"
}
],
"dataBindingPackages": [],
"compatibleVersions": [],
"dependencies": [
{
"name": "host",
"versionCode": 7359206
}
]
}`

func Test_getDepFeature(t *testing.T) {
	Convey("getDepFeature", t, func() {
		Convey("tribe feature json", func() {
			newJson := new(map[string]interface{})
			err := json.Unmarshal([]byte(newJsonStr), newJson)
			if err != nil {
				fmt.Printf("%v", err)
			}
			So(err, ShouldBeNil)
			feature := getDepFeature(*newJson)
			So(feature, ShouldEqual, "infra")
		})

		Convey("old json", func() {
			oldJson := new(map[string]interface{})
			err1 := json.Unmarshal([]byte(oldJsonStr), oldJson)
			if err1 != nil {
				fmt.Printf("%v", err1)
			}
			So(err1, ShouldBeNil)
			feature1 := getDepFeature(*oldJson)
			So(feature1, ShouldEqual, "default")
		})

	})
}
