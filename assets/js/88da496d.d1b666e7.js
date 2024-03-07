"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[3525],{9822:(e,t,s)=>{s.r(t),s.d(t,{assets:()=>c,contentTitle:()=>a,default:()=>l,frontMatter:()=>r,metadata:()=>i,toc:()=>m});var n=s(4848),o=s(8453);const r={id:"mattermost-integration_guide",title:"Mattermost Integration",sidebar_position:7},a=void 0,i={id:"guides/mattermost-integration_guide",title:"Mattermost Integration",description:"Overview",source:"@site/docs/guides/mattermost-integration_guide.md",sourceDirName:"guides",slug:"/guides/mattermost-integration_guide",permalink:"/argo-messaging/docs/guides/mattermost-integration_guide",draft:!1,unlisted:!1,tags:[],version:"current",sidebarPosition:7,frontMatter:{id:"mattermost-integration_guide",title:"Mattermost Integration",sidebar_position:7},sidebar:"tutorialSidebar",previous:{title:"Metrics Guide",permalink:"/argo-messaging/docs/guides/guide_metrics"},next:{title:"API Calls",permalink:"/argo-messaging/docs/category/api-calls"}},c={},m=[{value:"Overview",id:"overview",level:2},{value:"Mattermost Configuration",id:"mattermost-configuration",level:4},{value:"Subscription Configuration",id:"subscription-configuration",level:4},{value:"Reformat Messages Example",id:"reformat-messages-example",level:2}];function d(e){const t={a:"a",code:"code",h2:"h2",h4:"h4",li:"li",p:"p",pre:"pre",strong:"strong",ul:"ul",...(0,o.R)(),...e.components};return(0,n.jsxs)(n.Fragment,{children:[(0,n.jsx)(t.h2,{id:"overview",children:"Overview"}),"\n",(0,n.jsx)(t.p,{children:"Push enabled subscriptions provide us with the functionality to\nforward messages to mattermost channels via mattermost webhooks."}),"\n",(0,n.jsx)(t.h4,{id:"mattermost-configuration",children:"Mattermost Configuration"}),"\n",(0,n.jsxs)(t.p,{children:["Refer to this guide on how to set up your mattermost webhook.\n",(0,n.jsx)(t.a,{href:"https://mattermost.com/blog/mattermost-integrations-incoming-webhooks/",children:"https://mattermost.com/blog/mattermost-integrations-incoming-webhooks/"})]}),"\n",(0,n.jsx)(t.h4,{id:"subscription-configuration",children:"Subscription Configuration"}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-json",children:'{\n  "topic": "projects/example/topics/alarms-reformat-mattermost-topic",\n  "pushConfig": {\n    "type": "mattermost",\n    "maxMessages": 1,\n    "retryPolicy": {\n      "type": "linear",\n      "period": 3000\n    },\n    "mattermostUrl": "https://example.com/hooks/z5xjq7hzn7yobnjhthrh4q6oxw",\n    "mattermostUsername": "bot argo",\n    "mattermostChannel": "monitoring-alarms",\n    "base64Decode": true\n  }\n}\n'})}),"\n",(0,n.jsxs)(t.ul,{children:["\n",(0,n.jsxs)(t.li,{children:["\n",(0,n.jsxs)(t.p,{children:[(0,n.jsx)(t.code,{children:"mattermostUrl"}),": Is the webhook url that will be generated through\nthe integrations tab of the mattermost UI."]}),"\n"]}),"\n",(0,n.jsxs)(t.li,{children:["\n",(0,n.jsxs)(t.p,{children:[(0,n.jsx)(t.code,{children:"mattermostUsername"}),": Is the username that will be displayed alongside\nthe forwarded messages."]}),"\n"]}),"\n",(0,n.jsxs)(t.li,{children:["\n",(0,n.jsxs)(t.p,{children:[(0,n.jsx)(t.code,{children:"mattermostChannel"}),": Is the channel that the messages will be forwarded to."]}),"\n"]}),"\n",(0,n.jsxs)(t.li,{children:["\n",(0,n.jsxs)(t.p,{children:[(0,n.jsx)(t.code,{children:"base64Decode"}),": Messages in AMS should be base64 encoded.This flag allows a subscription\nto know if the the messages should be first decoded before being pushed\nto the remote destination.\nRefer to the following guides to better understand push enabled subscriptions\nand how to use them."]}),"\n"]}),"\n"]}),"\n",(0,n.jsx)(t.p,{children:(0,n.jsx)(t.a,{href:"http://argoeu.github.io/argo-messaging/openapi/explore#/Subscriptions/put_projects__PROJECT__subscriptions__SUBSCRIPTION_",children:"Swagger Create Subscription"})}),"\n",(0,n.jsx)(t.p,{children:(0,n.jsx)(t.a,{href:"http://argoeu.github.io/argo-messaging/docs/api_advanced/api_subscriptions#push-enabled-subscriptions",children:"Push Enabled Subscriptions"})}),"\n",(0,n.jsx)(t.h2,{id:"reformat-messages-example",children:"Reformat Messages Example"}),"\n",(0,n.jsx)(t.p,{children:"In some cases, a topic that has some raw messages, but we first\nwant to process them and reformat them, before pushing to mattermost,\nor reusing them for any other activity.\nIn order to achieve this we need to consume from the topic's subscription\nand republish them to another topic after the messages have been processes.\nWe then attach a push enabled subscription to the topic with the\nreformatted messages."}),"\n",(0,n.jsx)(t.p,{children:"The following snipper shows this kind of functionality."}),"\n",(0,n.jsxs)(t.p,{children:[(0,n.jsx)(t.strong,{children:"NOTE:"})," Implement your own ",(0,n.jsx)(t.code,{children:"format_message()"})," function to\ntransform messages to the desired format. The function accepts the\noriginal message decoded as input, and returns the formatted string."]}),"\n",(0,n.jsx)(t.pre,{children:(0,n.jsx)(t.code,{className:"language-python",children:'\n    # set up the ams client\n    ams_host = "{0}:{1}".format(args.host, str(args.port))\n    LOGGER.info("Setting up AMS client for host {0} and project: {1}".format(ams_host, args.project))\n    ams = ArgoMessagingService(endpoint=ams_host, project=args.project, token=args.token)\n\n    while True:\n        try:\n            # consume alerts\n            consumed_messages = ams.pull_sub(sub=args.sub, return_immediately=True, verify=args.verify)\n            if len(consumed_messages) == 0:\n                time.sleep(args.interval)\n                continue\n            payload = consumed_messages[0][1].get_data()\n            ack_id = consumed_messages[0][0]\n\n            # if we can\'t parse the message body we should ack the message and move to the next\n            try:\n                payload = json.loads(payload)\n                LOGGER.info("Examining new message {0} . . .".format(ack_id))\n\n                # skip messages that don\'t have a type of \'endpoint\' or \'group\'\n                if "type" not in payload or (payload["type"] != \'endpoint\' and payload["type"] != \'group\'):\n                    LOGGER.info("Skipping message {0} with wrong payload . . .".format(ack_id))\n                    try:\n                        ams.ack_sub(sub=args.sub, ids=[ack_id], verify=args.verify)\n                        continue\n                    except AmsException as e:\n                        LOGGER.error("Could not skip message {0}.{1}".format(ack_id, str(e)))\n                        continue\n            except Exception as e:\n                LOGGER.error("Cannot parse payload for message {0}.{1}.Skipping . . .".format(ack_id, str(e)))\n                try:\n                    ams.ack_sub(sub=args.sub, ids=[ack_id], verify=args.verify)\n                    continue\n                except AmsException as e:\n                    LOGGER.error("Could not skip message {0}.{1}".format(ack_id, str(e)))\n                    continue\n\n            # format and publish the new message\n            formatted_message = format_message(payload)\n            try:\n                ams.publish(topic=args.topic, msg=[AmsMessage(data=formatted_message)], verify=args.verify)\n            except AmsException as e:\n                LOGGER.error("Could not publish to topic.{0}".format(str(e)))\n                continue\n\n            # ack the original alert\n            try:\n                ams.ack_sub(sub=args.sub, ids=[ack_id], verify=args.verify)\n            except AmsException as e:\n                LOGGER.error("Could not ack original alert {0}.{1}".format(ack_id, str(e)))\n        except AmsException as e:\n            LOGGER.error("Cannot pull from subscription.{0}".format(str(e)))\n\n        time.sleep(args.interval)\n'})})]})}function l(e={}){const{wrapper:t}={...(0,o.R)(),...e.components};return t?(0,n.jsx)(t,{...e,children:(0,n.jsx)(d,{...e})}):d(e)}},8453:(e,t,s)=>{s.d(t,{R:()=>a,x:()=>i});var n=s(6540);const o={},r=n.createContext(o);function a(e){const t=n.useContext(r);return n.useMemo((function(){return"function"==typeof e?e(t):{...t,...e}}),[t,e])}function i(e){let t;return t=e.disableParentContext?"function"==typeof e.components?e.components(o):e.components||o:a(e.components),n.createElement(r.Provider,{value:t},e.children)}}}]);