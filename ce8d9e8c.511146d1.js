(window.webpackJsonp=window.webpackJsonp||[]).push([[25],{81:function(e,t,s){"use strict";s.r(t),s.d(t,"frontMatter",(function(){return o})),s.d(t,"metadata",(function(){return a})),s.d(t,"rightToc",(function(){return c})),s.d(t,"default",(function(){return p}));var n=s(2),i=s(6),r=(s(0),s(90)),o={id:"replaying_guide",title:"Receiving messages using Push"},a={unversionedId:"replaying_guide",id:"replaying_guide",isDocsHomePage:!1,title:"Receiving messages using Push",description:"Subscriptions\u2019 messages can still be accessed despite the fact that a Subscriber might have acknowledged them.",source:"@site/docs/replaying_guide.md",permalink:"/argo-messaging/docs/replaying_guide"},c=[{value:"Before you start",id:"before-you-start",children:[]},{value:"Get Subscription&#39;s offsets",id:"get-subscriptions-offsets",children:[]},{value:"Get Subscription&#39;s offsets by timestamp",id:"get-subscriptions-offsets-by-timestamp",children:[]},{value:"Move Subscription&#39;s offsets",id:"move-subscriptions-offsets",children:[]}],u={rightToc:c};function p(e){var t=e.components,s=Object(i.a)(e,["components"]);return Object(r.b)("wrapper",Object(n.a)({},u,s,{components:t,mdxType:"MDXLayout"}),Object(r.b)("p",null,"Subscriptions\u2019 messages can still be accessed despite the fact that a Subscriber might have acknowledged them."),Object(r.b)("p",null,"This functionality is possible through the subscription\u2019s offset modification.Each subscriptions holds three indices(offsets) that describe the messages that is providing (max:300, min:220, current: 288)."),Object(r.b)("p",null,"Whenever a message is acknowledged the current offset is incremented, indicating to the subscriber that the next message is available for consumption."),Object(r.b)("p",null,"In addition AMS provides the subscriber with the ability to seek offsets for a specific timestamp,the API will provide the closest possible offset it can find for the provided timestamp."),Object(r.b)("p",null,"Now that the subscriber has managed to retrieve the subscription\u2019s offset, we can use the modifyOffset api call  to move the indices around and re-consume/replay a subscription\u2019s messages."),Object(r.b)("p",null,"For example, if we have the offsets(max:300, min:220, current: 288), moving the current offset to 285, will allow the subscriber to again consume the messages ","[286,287,288]","."),Object(r.b)("p",null,"Last but not least, it is important to note that a message is being kept available through the AMS api for 7 days. After that time period has passed, it is no longer available and no offset can access it."),Object(r.b)("h2",{id:"before-you-start"},"Before you start"),Object(r.b)("p",null,"In order to get an account on the ARGO Messaging Service, submit a request through the ",Object(r.b)("a",Object(n.a)({parentName:"p"},{href:"https://docs.google.com/forms/d/e/1FAIpQLScfMCYPkUqUa5lT046RK1yCR4yn6M96WbgD5DMlNJ-zRFHSRA/viewform"}),"ARGO Messaging Service account form")),Object(r.b)("p",null,"Upon account approval, you will receive information via e-mail about your new project along with an API token."),Object(r.b)("h2",{id:"get-subscriptions-offsets"},"Get Subscription's offsets"),Object(r.b)("p",null,"A subscription\u2019s offsets can be accessed through the API using the following http call ",Object(r.b)("a",Object(n.a)({parentName:"p"},{href:"https://argoeu.github.io/argo-messaging/docs/api_subscriptions#get-get-offsets."}),"Get Subscription's offsets")),Object(r.b)("h2",{id:"get-subscriptions-offsets-by-timestamp"},"Get Subscription's offsets by timestamp"),Object(r.b)("p",null,"The following http call gives access to the aforementioned functionality ",Object(r.b)("a",Object(n.a)({parentName:"p"},{href:"https://argoeu.github.io/argo-messaging/docs/api_subscriptions#get-get-offset-by-timestamp"}),"Get Subscription's offsets by timestamp"),"."),Object(r.b)("h2",{id:"move-subscriptions-offsets"},"Move Subscription's offsets"),Object(r.b)("p",null,"The following http call gives access to the modifyOffset api call ",Object(r.b)("a",Object(n.a)({parentName:"p"},{href:"https://argoeu.github.io/argo-messaging/docs/api_subscriptions#post-modify-offsets"}),"Move Subscription's offsets")," to move the indices around and re-consume/replay a subscription\u2019s messages."))}p.isMDXComponent=!0},90:function(e,t,s){"use strict";s.d(t,"a",(function(){return f})),s.d(t,"b",(function(){return g}));var n=s(0),i=s.n(n);function r(e,t,s){return t in e?Object.defineProperty(e,t,{value:s,enumerable:!0,configurable:!0,writable:!0}):e[t]=s,e}function o(e,t){var s=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);t&&(n=n.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),s.push.apply(s,n)}return s}function a(e){for(var t=1;t<arguments.length;t++){var s=null!=arguments[t]?arguments[t]:{};t%2?o(Object(s),!0).forEach((function(t){r(e,t,s[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(s)):o(Object(s)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(s,t))}))}return e}function c(e,t){if(null==e)return{};var s,n,i=function(e,t){if(null==e)return{};var s,n,i={},r=Object.keys(e);for(n=0;n<r.length;n++)s=r[n],t.indexOf(s)>=0||(i[s]=e[s]);return i}(e,t);if(Object.getOwnPropertySymbols){var r=Object.getOwnPropertySymbols(e);for(n=0;n<r.length;n++)s=r[n],t.indexOf(s)>=0||Object.prototype.propertyIsEnumerable.call(e,s)&&(i[s]=e[s])}return i}var u=i.a.createContext({}),p=function(e){var t=i.a.useContext(u),s=t;return e&&(s="function"==typeof e?e(t):a(a({},t),e)),s},f=function(e){var t=p(e.components);return i.a.createElement(u.Provider,{value:t},e.children)},l={inlineCode:"code",wrapper:function(e){var t=e.children;return i.a.createElement(i.a.Fragment,{},t)}},b=i.a.forwardRef((function(e,t){var s=e.components,n=e.mdxType,r=e.originalType,o=e.parentName,u=c(e,["components","mdxType","originalType","parentName"]),f=p(s),b=n,g=f["".concat(o,".").concat(b)]||f[b]||l[b]||r;return s?i.a.createElement(g,a(a({ref:t},u),{},{components:s})):i.a.createElement(g,a({ref:t},u))}));function g(e,t){var s=arguments,n=t&&t.mdxType;if("string"==typeof e||n){var r=s.length,o=new Array(r);o[0]=b;var a={};for(var c in t)hasOwnProperty.call(t,c)&&(a[c]=t[c]);a.originalType=e,a.mdxType="string"==typeof e?e:n,o[1]=a;for(var u=2;u<r;u++)o[u]=s[u];return i.a.createElement.apply(null,o)}return i.a.createElement.apply(null,s)}b.displayName="MDXCreateElement"}}]);