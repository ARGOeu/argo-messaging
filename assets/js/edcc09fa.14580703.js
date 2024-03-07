"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[9464],{8974:(e,s,r)=>{r.r(s),r.d(s,{assets:()=>c,contentTitle:()=>i,default:()=>d,frontMatter:()=>a,metadata:()=>o,toc:()=>l});var n=r(4848),t=r(8453);const a={id:"api_metrics",title:"API Operational Metrics",sidebar_position:6},i=void 0,o={id:"api_advanced/api_metrics",title:"API Operational Metrics",description:"Operational Metrics include metrics related to the CPU or memory usage of the ams nodes",source:"@site/docs/api_advanced/api_metrics.md",sourceDirName:"api_advanced",slug:"/api_advanced/api_metrics",permalink:"/argo-messaging/docs/api_advanced/api_metrics",draft:!1,unlisted:!1,tags:[],version:"current",sidebarPosition:6,frontMatter:{id:"api_metrics",title:"API Operational Metrics",sidebar_position:6},sidebar:"tutorialSidebar",previous:{title:"Subscriptions",permalink:"/argo-messaging/docs/api_advanced/api_subscriptions"},next:{title:"Schemas",permalink:"/argo-messaging/docs/api_advanced/api_schemas"}},c={},l=[{value:"[GET] Get Operational Metrics",id:"get-get-operational-metrics",level:2},{value:"Request",id:"request",level:3},{value:"Example request",id:"example-request",level:3},{value:"Responses",id:"responses",level:3},{value:"Errors",id:"errors",level:3},{value:"[GET] Get Health status",id:"get-get-health-status",level:2},{value:"Request",id:"request-1",level:3},{value:"Example request",id:"example-request-1",level:3},{value:"Responses",id:"responses-1",level:3},{value:"Errors",id:"errors-1",level:3},{value:"[GET] Get a VA Report",id:"get-get-a-va-report",level:2},{value:"Request",id:"request-2",level:3},{value:"URL parameters",id:"url-parameters",level:3},{value:"Example request",id:"example-request-2",level:3},{value:"Example request with URL parameters",id:"example-request-with-url-parameters",level:3},{value:"Responses",id:"responses-2",level:3},{value:"Errors",id:"errors-2",level:3},{value:"[GET] Get User usage report",id:"get-get-user-usage-report",level:2},{value:"Request",id:"request-3",level:3},{value:"URL parameters",id:"url-parameters-1",level:3},{value:"Example request",id:"example-request-3",level:3},{value:"Example request with URL parameters",id:"example-request-with-url-parameters-1",level:3},{value:"Responses",id:"responses-3",level:3},{value:"Errors",id:"errors-3",level:3}];function p(e){const s={a:"a",code:"code",h2:"h2",h3:"h3",p:"p",pre:"pre",...(0,t.R)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(s.p,{children:"Operational Metrics include metrics related to the CPU or memory usage of the ams nodes"}),"\n",(0,n.jsx)(s.h2,{id:"get-get-operational-metrics",children:"[GET] Get Operational Metrics"}),"\n",(0,n.jsx)(s.p,{children:"This request gets a list of operational metrics for the specific ams service"}),"\n",(0,n.jsx)(s.h3,{id:"request",children:"Request"}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{children:'GET "/v1/metrics"\n'})}),"\n",(0,n.jsx)(s.h3,{id:"example-request",children:"Example request"}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-bash",children:'curl -H "Content-Type: application/json"\n "https://{URL}/v1/metrics?key=S3CR3T"\n'})}),"\n",(0,n.jsx)(s.h3,{id:"responses",children:"Responses"}),"\n",(0,n.jsx)(s.p,{children:"If successful, the response returns a list of related operational metrics"}),"\n",(0,n.jsxs)(s.p,{children:["Success Response\n",(0,n.jsx)(s.code,{children:"200 OK"})]}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-json",children:'{\n   "metrics": [\n      {\n         "metric": "ams_node.cpu_usage",\n         "metric_type": "percentage",\n         "value_type": "float64",\n         "resource_type": "ams_node",\n         "resource_name": "host.foo",\n         "timeseries": [\n            {\n               "timestamp": "2017-07-04T10:18:07Z",\n               "value": 0.2\n            }\n         ],\n         "description": "Percentage value that displays the CPU usage of ams service in the specific node"\n      },\n      {\n         "metric": "ams_node.memory_usage",\n         "metric_type": "percentage",\n         "value_type": "float64",\n         "resource_type": "ams_node",\n         "resource_name": "host.foo",\n         "timeseries": [\n            {\n               "timestamp": "2017-07-04T10:18:07Z",\n               "value": 0.1\n            }\n         ],\n         "description": "Percentage value that displays the Memory usage of ams service in the specific node"\n      }\n   ]\n}\n\n'})}),"\n",(0,n.jsx)(s.h3,{id:"errors",children:"Errors"}),"\n",(0,n.jsxs)(s.p,{children:["Please refer to section ",(0,n.jsx)(s.a,{href:"/argo-messaging/docs/api_basic/api_errors",children:"Errors"})," to see all possible Errors"]}),"\n",(0,n.jsx)(s.h2,{id:"get-get-health-status",children:"[GET] Get Health status"}),"\n",(0,n.jsx)(s.h3,{id:"request-1",children:"Request"}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{children:'GET "/v1/status"\n'})}),"\n",(0,n.jsx)(s.h3,{id:"example-request-1",children:"Example request"}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-bash",children:'curl -H "Content-Type: application/json"\n "https://{URL}/v1/status"\n'})}),"\n",(0,n.jsx)(s.h3,{id:"responses-1",children:"Responses"}),"\n",(0,n.jsx)(s.p,{children:"If successful, the response returns the health status of the service"}),"\n",(0,n.jsxs)(s.p,{children:["Success Response\n",(0,n.jsx)(s.code,{children:"200 OK"})]}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-json",children:'{\n  "status": "ok",\n  "push_servers": [\n    {\n      "endpoint": "localhost:5555",\n      "status": "Success: SERVING"\n    }\n  ]\n}\n'})}),"\n",(0,n.jsx)(s.h3,{id:"errors-1",children:"Errors"}),"\n",(0,n.jsxs)(s.p,{children:["Please refer to section ",(0,n.jsx)(s.a,{href:"/argo-messaging/docs/api_basic/api_errors",children:"Errors"})," to see all possible Errors"]}),"\n",(0,n.jsx)(s.h2,{id:"get-get-a-va-report",children:"[GET] Get a VA Report"}),"\n",(0,n.jsxs)(s.p,{children:["This request returns the total amount of messages per project for the given time window. The number of messages\nis calculated using the ",(0,n.jsx)(s.code,{children:"daily message count"})," for each one of the project's topics."]}),"\n",(0,n.jsx)(s.h3,{id:"request-2",children:"Request"}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{children:'GET "/v1/metrics/daily-message-average"\n\n'})}),"\n",(0,n.jsx)(s.h3,{id:"url-parameters",children:"URL parameters"}),"\n",(0,n.jsxs)(s.p,{children:[(0,n.jsx)(s.code,{children:"start_date"}),": start date for querying projects topics daily message count(optional), default value is the start unix time\n",(0,n.jsx)(s.code,{children:"end_date"}),": start date for querying projects topics daily message count(optional), default is the time of the api call\n",(0,n.jsx)(s.code,{children:"projects"}),": which projects to include to the query(optional), default is all registered projects"]}),"\n",(0,n.jsx)(s.h3,{id:"example-request-2",children:"Example request"}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-bash",children:'curl -H "Content-Type: application/json"\n "https://{URL}/v1/metrics/va_metrics"\n'})}),"\n",(0,n.jsx)(s.h3,{id:"example-request-with-url-parameters",children:"Example request with URL parameters"}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-bash",children:'curl -H "Content-Type: application/json"\n "https://{URL}/v1/metrics/va_metrics?start_date=2019-03-01&end_date=2019-07-24&projects=ARGO,ARGO-2"\n'})}),"\n",(0,n.jsx)(s.h3,{id:"responses-2",children:"Responses"}),"\n",(0,n.jsx)(s.p,{children:"If successful, the response returns the total amount of messages per project for the given time window"}),"\n",(0,n.jsxs)(s.p,{children:["Success Response\n",(0,n.jsx)(s.code,{children:"200 OK"})]}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-json",children:'{\n    "projects": [\n        {\n            "project": "ARGO-2",\n            "message_count": 8,\n            "average_daily_messages": 2\n        },\n        {\n            "project": "ARGO",\n            "message_count": 25669,\n            "average_daily_messages": 120\n        }\n    ],\n    "total_message_count": 25677,\n    "average_daily_messages": 122\n}\n'})}),"\n",(0,n.jsx)(s.h3,{id:"errors-2",children:"Errors"}),"\n",(0,n.jsxs)(s.p,{children:["Please refer to section ",(0,n.jsx)(s.a,{href:"/argo-messaging/docs/api_basic/api_errors",children:"Errors"})," to see all possible Errors"]}),"\n",(0,n.jsx)(s.h2,{id:"get-get-user-usage-report",children:"[GET] Get User usage report"}),"\n",(0,n.jsx)(s.p,{children:"This is a combination of the va_metrics and the operational_metrics\napi calls.The user will receive data for all of the projects that has\nthe project_admin role alongisde the operational metrics of the service."}),"\n",(0,n.jsx)(s.h3,{id:"request-3",children:"Request"}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{children:'GET "/v1/users/usageReport"\n\n'})}),"\n",(0,n.jsx)(s.h3,{id:"url-parameters-1",children:"URL parameters"}),"\n",(0,n.jsxs)(s.p,{children:[(0,n.jsx)(s.code,{children:"start_date"}),": start date for querying projects topics daily message count(optional), default value is the start unix time\n",(0,n.jsx)(s.code,{children:"end_date"}),": start date for querying projects topics daily message count(optional), default is the time of the api call\n",(0,n.jsx)(s.code,{children:"projects"}),": which projects to include to the query(optional), default is all registered projects"]}),"\n",(0,n.jsx)(s.h3,{id:"example-request-3",children:"Example request"}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-bash",children:'curl -H "Content-Type: application/json"\n "https://{URL}/v1/users/usageReport"\n'})}),"\n",(0,n.jsx)(s.h3,{id:"example-request-with-url-parameters-1",children:"Example request with URL parameters"}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-bash",children:'curl -H "Content-Type: application/json"\n "https://{URL}/v1/users/usageReport?start_date=2019-03-01&end_date=2019-07-24&projects=ARGO,ARGO-2"\n'})}),"\n",(0,n.jsx)(s.h3,{id:"responses-3",children:"Responses"}),"\n",(0,n.jsxs)(s.p,{children:["Success Response\n",(0,n.jsx)(s.code,{children:"200 OK"})]}),"\n",(0,n.jsx)(s.pre,{children:(0,n.jsx)(s.code,{className:"language-json",children:'{\n    "va_metrics": {\n        "projects_metrics": {\n            "projects": [\n                {\n                    "project": "e2epush",\n                    "message_count": 27,\n                    "average_daily_messages": 0.03,\n                    "topics_count": 3,\n                    "subscriptions_count": 6,\n                    "users_count": 0\n                }\n            ],\n            "total_message_count": 27,\n            "average_daily_messages": 0.03\n        },\n        "total_users_count": 0,\n        "total_topics_count": 3,\n        "total_subscriptions_count": 6\n    },\n    "operational_metrics": {\n        "metrics": [\n            {\n                "metric": "ams_node.cpu_usage",\n                "metric_type": "percentage",\n                "value_type": "float64",\n                "resource_type": "ams_node",\n                "resource_name": "test-MBP",\n                "timeseries": [\n                    {\n                        "timestamp": "2022-09-13T09:39:56Z",\n                        "value": 0\n                    }\n                ],\n                "description": "Percentage value that displays the CPU usage of ams service in the specific node"\n            },\n            {\n                "metric": "ams_node.memory_usage",\n                "metric_type": "percentage",\n                "value_type": "float64",\n                "resource_type": "ams_node",\n                "resource_name": "test-MBP",\n                "timeseries": [\n                    {\n                        "timestamp": "2022-09-13T09:39:56Z",\n                        "value": 0.1\n                    }\n                ],\n                "description": "Percentage value that displays the Memory usage of ams service in the specific node"\n            },\n            {\n                "metric": "ams_node.cpu_usage",\n                "metric_type": "percentage",\n                "value_type": "float64",\n                "resource_type": "ams_node",\n                "resource_name": "4",\n                "timeseries": [\n                    {\n                        "timestamp": "2022-09-13T09:39:56Z",\n                        "value": 0\n                    }\n                ],\n                "description": "Percentage value that displays the CPU usage of ams service in the specific node"\n            },\n            {\n                "metric": "ams_node.memory_usage",\n                "metric_type": "percentage",\n                "value_type": "float64",\n                "resource_type": "ams_node",\n                "resource_name": "4",\n                "timeseries": [\n                    {\n                        "timestamp": "2022-09-13T09:39:56Z",\n                        "value": 0.1\n                    }\n                ],\n                "description": "Percentage value that displays the Memory usage of ams service in the specific node"\n            }\n        ]\n    }\n}\n'})}),"\n",(0,n.jsx)(s.h3,{id:"errors-3",children:"Errors"}),"\n",(0,n.jsxs)(s.p,{children:["Please refer to section ",(0,n.jsx)(s.a,{href:"/argo-messaging/docs/api_basic/api_errors",children:"Errors"})," to see all possible Errors"]})]})}function d(e={}){const{wrapper:s}={...(0,t.R)(),...e.components};return s?(0,n.jsx)(s,{...e,children:(0,n.jsx)(p,{...e})}):p(e)}},8453:(e,s,r)=>{r.d(s,{R:()=>i,x:()=>o});var n=r(6540);const t={},a=n.createContext(t);function i(e){const s=n.useContext(a);return n.useMemo((function(){return"function"==typeof e?e(s):{...s,...e}}),[s,e])}function o(e){let s;return s=e.disableParentContext?"function"==typeof e.components?e.components(t):e.components||t:i(e.components),n.createElement(a.Provider,{value:s},e.children)}}}]);