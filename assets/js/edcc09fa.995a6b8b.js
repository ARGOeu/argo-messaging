"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[2211],{3905:(e,t,r)=>{r.d(t,{Zo:()=>c,kt:()=>d});var a=r(7294);function s(e,t,r){return t in e?Object.defineProperty(e,t,{value:r,enumerable:!0,configurable:!0,writable:!0}):e[t]=r,e}function n(e,t){var r=Object.keys(e);if(Object.getOwnPropertySymbols){var a=Object.getOwnPropertySymbols(e);t&&(a=a.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),r.push.apply(r,a)}return r}function i(e){for(var t=1;t<arguments.length;t++){var r=null!=arguments[t]?arguments[t]:{};t%2?n(Object(r),!0).forEach((function(t){s(e,t,r[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(r)):n(Object(r)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(r,t))}))}return e}function o(e,t){if(null==e)return{};var r,a,s=function(e,t){if(null==e)return{};var r,a,s={},n=Object.keys(e);for(a=0;a<n.length;a++)r=n[a],t.indexOf(r)>=0||(s[r]=e[r]);return s}(e,t);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);for(a=0;a<n.length;a++)r=n[a],t.indexOf(r)>=0||Object.prototype.propertyIsEnumerable.call(e,r)&&(s[r]=e[r])}return s}var l=a.createContext({}),p=function(e){var t=a.useContext(l),r=t;return e&&(r="function"==typeof e?e(t):i(i({},t),e)),r},c=function(e){var t=p(e.components);return a.createElement(l.Provider,{value:t},e.children)},u={inlineCode:"code",wrapper:function(e){var t=e.children;return a.createElement(a.Fragment,{},t)}},m=a.forwardRef((function(e,t){var r=e.components,s=e.mdxType,n=e.originalType,l=e.parentName,c=o(e,["components","mdxType","originalType","parentName"]),m=p(r),d=s,g=m["".concat(l,".").concat(d)]||m[d]||u[d]||n;return r?a.createElement(g,i(i({ref:t},c),{},{components:r})):a.createElement(g,i({ref:t},c))}));function d(e,t){var r=arguments,s=t&&t.mdxType;if("string"==typeof e||s){var n=r.length,i=new Array(n);i[0]=m;var o={};for(var l in t)hasOwnProperty.call(t,l)&&(o[l]=t[l]);o.originalType=e,o.mdxType="string"==typeof e?e:s,i[1]=o;for(var p=2;p<n;p++)i[p]=r[p];return a.createElement.apply(null,i)}return a.createElement.apply(null,r)}m.displayName="MDXCreateElement"},550:(e,t,r)=>{r.r(t),r.d(t,{assets:()=>l,contentTitle:()=>i,default:()=>u,frontMatter:()=>n,metadata:()=>o,toc:()=>p});var a=r(7462),s=(r(7294),r(3905));const n={id:"api_metrics",title:"API Operational Metrics",sidebar_position:6},i=void 0,o={unversionedId:"api_advanced/api_metrics",id:"api_advanced/api_metrics",title:"API Operational Metrics",description:"Operational Metrics include metrics related to the CPU or memory usage of the ams nodes",source:"@site/docs/api_advanced/api_metrics.md",sourceDirName:"api_advanced",slug:"/api_advanced/api_metrics",permalink:"/argo-messaging/docs/api_advanced/api_metrics",draft:!1,tags:[],version:"current",sidebarPosition:6,frontMatter:{id:"api_metrics",title:"API Operational Metrics",sidebar_position:6},sidebar:"tutorialSidebar",previous:{title:"Subscriptions",permalink:"/argo-messaging/docs/api_advanced/api_subscriptions"},next:{title:"Schemas",permalink:"/argo-messaging/docs/api_advanced/api_schemas"}},l={},p=[{value:"GET Get Operational Metrics",id:"get-get-operational-metrics",level:2},{value:"Request",id:"request",level:3},{value:"Example request",id:"example-request",level:3},{value:"Responses",id:"responses",level:3},{value:"Errors",id:"errors",level:3},{value:"GET Get Health status",id:"get-get-health-status",level:2},{value:"Request",id:"request-1",level:3},{value:"Example request",id:"example-request-1",level:3},{value:"Responses",id:"responses-1",level:3},{value:"Errors",id:"errors-1",level:3},{value:"GET Get Daily Message Average",id:"get-get-daily-message-average",level:2},{value:"Request",id:"request-2",level:3},{value:"URL parameters",id:"url-parameters",level:3},{value:"Example request",id:"example-request-2",level:3},{value:"Example request with URL parameters",id:"example-request-with-url-parameters",level:3},{value:"Responses",id:"responses-2",level:3},{value:"Errors",id:"errors-2",level:3}],c={toc:p};function u(e){let{components:t,...r}=e;return(0,s.kt)("wrapper",(0,a.Z)({},c,r,{components:t,mdxType:"MDXLayout"}),(0,s.kt)("p",null,"Operational Metrics include metrics related to the CPU or memory usage of the ams nodes"),(0,s.kt)("h2",{id:"get-get-operational-metrics"},"[GET]"," Get Operational Metrics"),(0,s.kt)("p",null,"This request gets a list of operational metrics for the specific ams service"),(0,s.kt)("h3",{id:"request"},"Request"),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre"},'GET "/v1/metrics"\n')),(0,s.kt)("h3",{id:"example-request"},"Example request"),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-bash"},'curl -H "Content-Type: application/json"\n "https://{URL}/v1/metrics?key=S3CR3T"\n')),(0,s.kt)("h3",{id:"responses"},"Responses"),(0,s.kt)("p",null,"If successful, the response returns a list of related operational metrics"),(0,s.kt)("p",null,"Success Response\n",(0,s.kt)("inlineCode",{parentName:"p"},"200 OK")),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-json"},'{\n   "metrics": [\n      {\n         "metric": "ams_node.cpu_usage",\n         "metric_type": "percentage",\n         "value_type": "float64",\n         "resource_type": "ams_node",\n         "resource_name": "host.foo",\n         "timeseries": [\n            {\n               "timestamp": "2017-07-04T10:18:07Z",\n               "value": 0.2\n            }\n         ],\n         "description": "Percentage value that displays the CPU usage of ams service in the specific node"\n      },\n      {\n         "metric": "ams_node.memory_usage",\n         "metric_type": "percentage",\n         "value_type": "float64",\n         "resource_type": "ams_node",\n         "resource_name": "host.foo",\n         "timeseries": [\n            {\n               "timestamp": "2017-07-04T10:18:07Z",\n               "value": 0.1\n            }\n         ],\n         "description": "Percentage value that displays the Memory usage of ams service in the specific node"\n      }\n   ]\n}\n\n')),(0,s.kt)("h3",{id:"errors"},"Errors"),(0,s.kt)("p",null,"Please refer to section ",(0,s.kt)("a",{parentName:"p",href:"/argo-messaging/docs/api_basic/api_errors"},"Errors")," to see all possible Errors"),(0,s.kt)("h2",{id:"get-get-health-status"},"[GET]"," Get Health status"),(0,s.kt)("h3",{id:"request-1"},"Request"),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre"},'GET "/v1/status"\n')),(0,s.kt)("h3",{id:"example-request-1"},"Example request"),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-bash"},'curl -H "Content-Type: application/json"\n "https://{URL}/v1/status"\n')),(0,s.kt)("h3",{id:"responses-1"},"Responses"),(0,s.kt)("p",null,"If successful, the response returns the health status of the service"),(0,s.kt)("p",null,"Success Response\n",(0,s.kt)("inlineCode",{parentName:"p"},"200 OK")),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-json"},'{\n  "status": "ok",\n  "push_servers": [\n    {\n      "endpoint": "localhost:5555",\n      "status": "Success: SERVING"\n    }\n  ]\n}\n')),(0,s.kt)("h3",{id:"errors-1"},"Errors"),(0,s.kt)("p",null,"Please refer to section ",(0,s.kt)("a",{parentName:"p",href:"/argo-messaging/docs/api_basic/api_errors"},"Errors")," to see all possible Errors"),(0,s.kt)("h2",{id:"get-get-daily-message-average"},"[GET]"," Get Daily Message Average"),(0,s.kt)("p",null,"This request returns the total amount of messages per project for the given time window. The number of messages\nis calculated using the ",(0,s.kt)("inlineCode",{parentName:"p"},"daily message count")," for each one of the project's topics."),(0,s.kt)("h3",{id:"request-2"},"Request"),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre"},'GET "/v1/metrics/daily-message-average"\n\n')),(0,s.kt)("h3",{id:"url-parameters"},"URL parameters"),(0,s.kt)("p",null,(0,s.kt)("inlineCode",{parentName:"p"},"start_date"),": start date for querying projects topics daily message count(optional), default value is the start unix time\n",(0,s.kt)("inlineCode",{parentName:"p"},"end_date"),": start date for querying projects topics daily message count(optional), default is the time of the api call\n",(0,s.kt)("inlineCode",{parentName:"p"},"projects"),": which projects to include to the query(optional), default is all registered projects"),(0,s.kt)("h3",{id:"example-request-2"},"Example request"),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-bash"},'curl -H "Content-Type: application/json"\n "https://{URL}/v1/metrics/daily-message-average"\n')),(0,s.kt)("h3",{id:"example-request-with-url-parameters"},"Example request with URL parameters"),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-bash"},'curl -H "Content-Type: application/json"\n "https://{URL}/v1/metrics/daily-message-average?start_date=2019-03-01&end_date=2019-07-24&projects=ARGO,ARGO-2"\n')),(0,s.kt)("h3",{id:"responses-2"},"Responses"),(0,s.kt)("p",null,"If successful, the response returns the total amount of messages per project for the given time window"),(0,s.kt)("p",null,"Success Response\n",(0,s.kt)("inlineCode",{parentName:"p"},"200 OK")),(0,s.kt)("pre",null,(0,s.kt)("code",{parentName:"pre",className:"language-json"},'{\n    "projects": [\n        {\n            "project": "ARGO-2",\n            "message_count": 8,\n            "average_daily_messages": 2\n        },\n        {\n            "project": "ARGO",\n            "message_count": 25669,\n            "average_daily_messages": 120\n        }\n    ],\n    "total_message_count": 25677,\n    "average_daily_messages": 122\n}\n')),(0,s.kt)("h3",{id:"errors-2"},"Errors"),(0,s.kt)("p",null,"Please refer to section ",(0,s.kt)("a",{parentName:"p",href:"/argo-messaging/docs/api_basic/api_errors"},"Errors")," to see all possible Errors"))}u.isMDXComponent=!0}}]);