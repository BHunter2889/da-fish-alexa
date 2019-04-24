package main

const (
	aplJson = `{
    "document": {
        "type": "APL",
        "version": "1.0",
        "theme": "dark",
        "import": [
            {
                "name": "alexa-layouts",
                "version": "1.0.0"
            },
            {
                "name": "alexa-styles",
                "version": "1.0.0"
            }
        ],
        "resources": [
			{
                "description": "Stock color for the dark theme",
                "when": "${viewport.theme == 'dark'}",
                "colors": {
                    "colorTextPrimary": "#f0f1ef"
                }
            },
            {
                "description": "Stock color for the light theme",
                "when": "${viewport.theme == 'light'}",
                "colors": {
                    "colorTextPrimary": "#151920"
                }
            },
			{
                "description": "Stock overlay color for the dark theme",
                "when": "${viewport.theme == 'dark'}",
                "colors": {
                    "colorBackgroundOverlay": "#00000050"
                }
            },
            {
                "description": "Stock overlay color for the light theme",
                "when": "${viewport.theme == 'light'}",
                "colors": {
					"colorBackgroundOverlay": "#FAFAFA77"
                }
            },
            {
                "description": "Standard font sizes",
                "dimensions": {
                    "textSizeBody": 48,
                    "textSizePrimary": 27,
                    "textSizeSecondary": 23,
                    "textSizeSecondaryHint": 25
                }
            },
            {
                "description": "Common spacing values",
                "dimensions": {
                    "spacingThin": 6,
                    "spacingSmall": 12,
                    "spacingMedium": 24,
                    "spacingLarge": 48,
                    "spacingExtraLarge": 72
                }
            },
            {
                "description": "Common margins and padding",
                "dimensions": {
                    "marginTop": 40,
                    "marginLeft": 60,
                    "marginRight": 60,
                    "marginBottom": 40
                }
            }
        ],
        "styles": {
            "textStyleBase": {
                "description": "Base font description; set color and core font family",
                "values": [
                    {
                        "color": "@colorTextPrimary",
                        "fontFamily": "Amazon Ember"
                    }
                ]
            },
            "textStyleBase0": {
                "description": "Thin version of basic font",
                "extend": "textStyleBase",
                "values": {
                    "fontWeight": "100"
                }
            },
            "textStyleBase1": {
                "description": "Light version of basic font",
                "extend": "textStyleBase",
                "values": {
                    "fontWeight": "300"
                }
            },
            "mixinBody": {
                "values": {
                    "fontSize": "@textSizeBody"
                }
            },
            "mixinPrimary": {
                "values": {
                    "fontSize": "@textSizePrimary"
                }
            },
            "mixinSecondary": {
                "values": {
                    "fontSize": "@textSizeSecondary"
                }
            },
            "textStylePrimary": {
                "extend": [
                    "textStyleBase1",
                    "mixinPrimary"
                ]
            },
            "textStyleSecondary": {
                "extend": [
                    "textStyleBase0",
                    "mixinSecondary"
                ]
            },
            "textStyleBody": {
                "extend": [
                    "textStyleBase1",
                    "mixinBody"
                ]
            },
            "textStyleSecondaryHint": {
                "values": {
                    "fontFamily": "Bookerly",
                    "fontStyle": "italic",
                    "fontSize": "@textSizeSecondaryHint",
                    "color": "@colorTextPrimary"
                }
            }
        },
        "layouts": {},
        "mainTemplate": {
            "description": "********* Full-screen background image **********",
            "parameters": [
                "payload"
            ],
            "items": [
                {
                    "when": "${viewport.shape == 'round'}",
                    "type": "Container",
                    "direction": "column",
					"height": "100vh",
                    "width": "100vw",
                    "items": [
                        {
                            "type": "Image",
                            "source": "${payload.bodyTemplate1Data.backgroundImage.sources[0].url}",
                            "overlayColor": "@colorBackgroundOverlay",
                            "position": "absolute",
                            "width": "100vw",
                            "height": "100vh",
                            "scale": "best-fill"
                        },
                        {
                            "type": "AlexaHeader",
                            "headerTitle": "${payload.bodyTemplate1Data.title}",
                            "headerAttributionImage": "${payload.bodyTemplate1Data.logoUrl}"
                        },
                        {
                            "type": "Container",
                            "grow": 1,
                            "paddingLeft": "@marginLeft",
                            "paddingRight": "@marginRight",
                            "paddingBottom": "@marginBottom"
                        }
                    ]
                },
                {
                    "type": "Container",
                    "height": "100vh",
                    "items": [
                        {
                            "type": "Image",
                            "source": "${payload.bodyTemplate1Data.backgroundImage.sources[0].url}",
                            "overlayColor": "@colorBackgroundOverlay",
                            "position": "absolute",
                            "width": "100vw",
                            "height": "100vh",
                            "scale": "best-fill"
                        },
                        {
                            "type": "AlexaHeader",
                            "headerTitle": "${payload.bodyTemplate1Data.title}",
                            "headerAttributionImage": "${payload.bodyTemplate1Data.logoUrl}"
                        },
                        {
                            "type": "Container",
                            "paddingLeft": "@marginLeft",
                            "paddingRight": "@marginRight",
                            "paddingBottom": "@marginBottom",
                            "items": [
                                {
                                    "type": "Text",
                                    "text": "${payload.bodyTemplate1Data.textContent.primaryText.text}",
                                    "fontSize": "@textSizeBody",
                                    "spacing": "@spacingSmall",
                                    "style": "textStyleBody"
                                }
                            ]
                        }
                    ]
                }
            ]
        }
    },
    "dataSources": {
        "bodyTemplate1Data": {
            "type": "object",
            "objectId": null,
            "backgroundImage": {
                "contentDescription": "Full-screen background image of a person fly fishing in a boat on a lake surrounded by dense evergreen tree foliage.",
                "smallSourceUrl": "https://s3.amazonaws.com/bugcaster-resources/fishing-canoe-1024.png",
                "largeSourceUrl": "https://s3.amazonaws.com/bugcaster-resources/fishing-canoe-1024.png",
                "sources": [
                    {
                        "url": "https://s3.amazonaws.com/bugcaster-resources/fishing-canoe-1024.png",
                        "size": "small",
                        "widthPixels": 0,
                        "heightPixels": 0
                    },
                    {
                        "url": "https://s3.amazonaws.com/bugcaster-resources/fishing-canoe-1024.png",
                        "size": "large",
                        "widthPixels": 0,
                        "heightPixels": 0
                    }
                ]
            },
            "title": "Today's Fishing Forecast - Next Two Hours",
            "textContent": {
                "primaryText": {
                    "type": "PlainText",
                    "text": "Text (or SSML) response goes Here. This Should Be at least long enough to fill a Spot device."
                }
            },
            "logoUrl": "https://s3.amazonaws.com/bugcaster-resources/bugcaster-logo.PNG"
        }
    }
}`
)