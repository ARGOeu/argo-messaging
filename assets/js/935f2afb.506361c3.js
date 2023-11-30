"use strict";(self.webpackChunkwebsite=self.webpackChunkwebsite||[]).push([[53],{1109:e=>{e.exports=JSON.parse('{"pluginId":"default","version":"current","label":"Next","banner":null,"badge":false,"className":"docs-version-current","isLast":true,"docsSidebars":{"tutorialSidebar":[{"type":"category","label":"What is AMS","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"AMS - The Service","href":"/argo-messaging/docs/","docId":"intro/what_is_ams"}],"href":"/argo-messaging/docs/category/what-is-ams"},{"type":"category","label":"General Concepts","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"Overview & Introduction","href":"/argo-messaging/docs/concepts/overview","docId":"concepts/overview"},{"type":"link","label":"Using Apache Kafka as a Backend Message system","href":"/argo-messaging/docs/concepts/msg_backend","docId":"concepts/msg_backend"},{"type":"link","label":"Data flow in Argo Messaging","href":"/argo-messaging/docs/concepts/msg_flow","docId":"concepts/msg_flow"},{"type":"link","label":"Authentication & Authorization","href":"/argo-messaging/docs/concepts/auth","docId":"concepts/auth"},{"type":"link","label":"Initial Project & User Management","href":"/argo-messaging/docs/concepts/projects_users","docId":"concepts/projects_users"}],"href":"/argo-messaging/docs/category/general-concepts"},{"type":"category","label":"Guides","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"Publisher Guide","href":"/argo-messaging/docs/guides/publisher","docId":"guides/publisher"},{"type":"link","label":"Subscriber Guide","href":"/argo-messaging/docs/guides/subscriber_guide","docId":"guides/subscriber_guide"},{"type":"link","label":"Receiving messages using Pull","href":"/argo-messaging/docs/guides/subscriber-pull_guide","docId":"guides/subscriber-pull_guide"},{"type":"link","label":"Receiving messages using Push","href":"/argo-messaging/docs/guides/subscriber-push_guide","docId":"guides/subscriber-push_guide"},{"type":"link","label":"Replaying Messages","href":"/argo-messaging/docs/guides/replaying_guide","docId":"guides/replaying_guide"},{"type":"link","label":"Metrics Guide","href":"/argo-messaging/docs/guides/guide_metrics","docId":"guides/guide_metrics"},{"type":"link","label":"Mattermost Integration","href":"/argo-messaging/docs/guides/mattermost-integration_guide","docId":"guides/mattermost-integration_guide"}],"href":"/argo-messaging/docs/category/guides"},{"type":"category","label":"API Calls","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"Authentication","href":"/argo-messaging/docs/api_advanced/api_auth","docId":"api_advanced/api_auth"},{"type":"link","label":"User Management","href":"/argo-messaging/docs/api_advanced/api_users","docId":"api_advanced/api_users"},{"type":"link","label":"Projects","href":"/argo-messaging/docs/api_advanced/api_projects","docId":"api_advanced/api_projects"},{"type":"link","label":"Topics","href":"/argo-messaging/docs/api_advanced/api_topics","docId":"api_advanced/api_topics"},{"type":"link","label":"Subscriptions","href":"/argo-messaging/docs/api_advanced/api_subscriptions","docId":"api_advanced/api_subscriptions"},{"type":"link","label":"API Operational Metrics","href":"/argo-messaging/docs/api_advanced/api_metrics","docId":"api_advanced/api_metrics"},{"type":"link","label":"Schemas","href":"/argo-messaging/docs/api_advanced/api_schemas","docId":"api_advanced/api_schemas"},{"type":"link","label":"Get API Version information","href":"/argo-messaging/docs/api_advanced/api_version","docId":"api_advanced/api_version"},{"type":"link","label":"User Registration","href":"/argo-messaging/docs/api_advanced/api_registrations","docId":"api_advanced/api_registrations"}],"href":"/argo-messaging/docs/category/api-calls"},{"type":"category","label":"Argo Messaging API","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"Service introduction and configuration","href":"/argo-messaging/docs/api_basic/api_intro","docId":"api_basic/api_intro"},{"type":"link","label":"API Errors","href":"/argo-messaging/docs/api_basic/api_errors","docId":"api_basic/api_errors"}],"href":"/argo-messaging/docs/category/argo-messaging-api"},{"type":"category","label":"How to use","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"How to use the service","href":"/argo-messaging/docs/howto/how_to_use","docId":"howto/how_to_use"},{"type":"link","label":"AMS Push Worker","href":"/argo-messaging/docs/howto/ams_push_worker","docId":"howto/ams_push_worker"}],"href":"/argo-messaging/docs/category/how-to-use"},{"type":"category","label":"Frequent Questions","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"Q & A - General Questions","href":"/argo-messaging/docs/faq/qa","docId":"faq/qa"},{"type":"link","label":"Q & A - Ruby specifics","href":"/argo-messaging/docs/faq/qa_ruby","docId":"faq/qa_ruby"}],"href":"/argo-messaging/docs/category/frequent-questions"},{"type":"category","label":"Policies","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"Terms of Use","href":"/argo-messaging/docs/policies/terms","docId":"policies/terms"},{"type":"link","label":"AMS Privacy Policy","href":"/argo-messaging/docs/policies/ams_privacy_policy","docId":"policies/ams_privacy_policy"},{"type":"link","label":"Technical and organisational measures (TOM)","href":"/argo-messaging/docs/policies/tom","docId":"policies/tom"}],"href":"/argo-messaging/docs/category/policies"},{"type":"category","label":"Use Cases","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"Use cases","href":"/argo-messaging/docs/use_cases/","docId":"use_cases/use_cases"}],"href":"/argo-messaging/docs/category/use-cases"},{"type":"category","label":"Training Material","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"Training Material","href":"/argo-messaging/docs/training_material/","docId":"training_material/training_material"}],"href":"/argo-messaging/docs/category/training-material"},{"type":"category","label":"Communication","collapsible":true,"collapsed":true,"items":[{"type":"link","label":"Communication Channels","href":"/argo-messaging/docs/communication/","docId":"communication/communication_channels"}],"href":"/argo-messaging/docs/category/communication"}]},"docs":{"api_advanced/api_auth":{"id":"api_advanced/api_auth","title":"Authentication","description":"Each user is authenticated by adding the url parameter ?key=T0K3N in each API request","sidebar":"tutorialSidebar"},"api_advanced/api_metrics":{"id":"api_advanced/api_metrics","title":"API Operational Metrics","description":"Operational Metrics include metrics related to the CPU or memory usage of the ams nodes","sidebar":"tutorialSidebar"},"api_advanced/api_projects":{"id":"api_advanced/api_projects","title":"Projects","description":"ARGO Messaging Service supports project entities as a basis of organizing and isolating groups of users & resources","sidebar":"tutorialSidebar"},"api_advanced/api_registrations":{"id":"api_advanced/api_registrations","title":"User Registration","description":"ARGO Messaging Service supports calls for registering users","sidebar":"tutorialSidebar"},"api_advanced/api_schemas":{"id":"api_advanced/api_schemas","title":"Schemas","description":"Schemas is a resource that works with topics by validating the published messages.","sidebar":"tutorialSidebar"},"api_advanced/api_subscriptions":{"id":"api_advanced/api_subscriptions","title":"Subscriptions","description":"[PUT] Manage Subscriptions - Create subscriptions","sidebar":"tutorialSidebar"},"api_advanced/api_topics":{"id":"api_advanced/api_topics","title":"Topics","description":"Topics are resources that can hold messages. Publishers (users/systems) can create topics on demand and name them (Usually with names that make sense and express the class of messages delivered in the topic).","sidebar":"tutorialSidebar"},"api_advanced/api_users":{"id":"api_advanced/api_users","title":"User Management","description":"ARGO Messaging Service supports calls for creating and modifying users","sidebar":"tutorialSidebar"},"api_advanced/api_version":{"id":"api_advanced/api_version","title":"Get API Version information","description":"This method can be used to retrieve api version information","sidebar":"tutorialSidebar"},"api_basic/api_errors":{"id":"api_basic/api_errors","title":"API Errors","description":"Errors","sidebar":"tutorialSidebar"},"api_basic/api_intro":{"id":"api_basic/api_intro","title":"Service introduction and configuration","description":"Introduction","sidebar":"tutorialSidebar"},"communication/communication_channels":{"id":"communication/communication_channels","title":"Communication Channels","description":"There are two ways you can initiate communication with the team behind the","sidebar":"tutorialSidebar"},"concepts/auth":{"id":"concepts/auth","title":"Authentication & Authorization","description":"Authentication is the process of determining the identity of a client, which is typically a user account. Authorization is the process of determining what permissions an authenticated identity has on a set of specified resources. In the Messaging API, there can be no authorization without authentication.","sidebar":"tutorialSidebar"},"concepts/msg_backend":{"id":"concepts/msg_backend","title":"Using Apache Kafka as a Backend Message system","description":"The ARGO Messaging API has been designed to rely on a generic Message Back-end Interface and use specific implementation of that interface for supporting different systems. Right now the first implementation for the messaging back-end relies on Apache Kafka as a distributed messaging system.","sidebar":"tutorialSidebar"},"concepts/msg_flow":{"id":"concepts/msg_flow","title":"Data flow in Argo Messaging","description":"The main steps of the messaging API:","sidebar":"tutorialSidebar"},"concepts/overview":{"id":"concepts/overview","title":"Overview & Introduction","description":"The Messaging Services is implemented as a Publish/Subscribe Service. Instead of focusing on a single Messaging API specification for handling the logic of publishing/subscribing to the broker network the API focuses on creating nodes of Publishers and Subscribers as a Service.","sidebar":"tutorialSidebar"},"concepts/projects_users":{"id":"concepts/projects_users","title":"Initial Project & User Management","description":"This document describes some of the more advanced setup you may need to do while configuring and deploying the ARGO Messaging Service.","sidebar":"tutorialSidebar"},"faq/qa":{"id":"faq/qa","title":"Q & A - General Questions","description":"Questions and answers based on problems encountered during implementation.","sidebar":"tutorialSidebar"},"faq/qa_ruby":{"id":"faq/qa_ruby","title":"Q & A - Ruby specifics","description":"Questions and answers based on problems encountered during implementation.","sidebar":"tutorialSidebar"},"guides/guide_metrics":{"id":"guides/guide_metrics","title":"Metrics Guide","description":"Project metrics:  If you want to see the specific project metrics please visit this page Project Metrics","sidebar":"tutorialSidebar"},"guides/mattermost-integration_guide":{"id":"guides/mattermost-integration_guide","title":"Mattermost Integration","description":"Overview","sidebar":"tutorialSidebar"},"guides/publisher":{"id":"guides/publisher","title":"Publisher Guide","description":"Publishers can send messages to named-channels called Topics.","sidebar":"tutorialSidebar"},"guides/replaying_guide":{"id":"guides/replaying_guide","title":"Replaying Messages","description":"Subscriptions\u2019 messages can still be accessed despite the fact that a Subscriber might have acknowledged them.","sidebar":"tutorialSidebar"},"guides/subscriber_guide":{"id":"guides/subscriber_guide","title":"Subscriber Guide","description":"Subscribers can read messages from named-channels called Subscriptions.  Each subscription can belong to a single topic. A topic though can have multiple subscriptions.","sidebar":"tutorialSidebar"},"guides/subscriber-pull_guide":{"id":"guides/subscriber-pull_guide","title":"Receiving messages using Pull","description":"Subscribers can read messages from named-channels called Subscriptions.  Each subscription can belong to a single topic. A topic though can have multiple subscriptions.","sidebar":"tutorialSidebar"},"guides/subscriber-push_guide":{"id":"guides/subscriber-push_guide","title":"Receiving messages using Push","description":"Subscribers can read messages from named-channels called Subscriptions.  Each subscription can belong to a single topic. A topic though can have multiple subscriptions.","sidebar":"tutorialSidebar"},"howto/ams_push_worker":{"id":"howto/ams_push_worker","title":"AMS Push Worker","description":"AMS Push worker (ver 0.1.0) is a command line utility that let\u2019s you simulate AMS push functionality by pulling messages from an actual AMS project/subscription and pushing them to an endpoint in your local development environment. It\u2019s written in go and it\u2019s packaged as a single binary with no dependencies.","sidebar":"tutorialSidebar"},"howto/how_to_use":{"id":"howto/how_to_use","title":"How to use the service","description":"Ideas","sidebar":"tutorialSidebar"},"intro/what_is_ams":{"id":"intro/what_is_ams","title":"AMS - The Service","description":"The ARGO Messaging Service (AMS)  is a Publish/Subscribe Service, which implements the Google PubSub protocol. Instead of focusing on a single Messaging API specification for handling the logic of publishing/subscribing to the broker network the API focuses on creating nodes of Publishers and Subscribers as a Service. It provides an HTTP API that enables Users/Systems to implement message oriented service using the Publish/Subscribe Model over plain HTTP.","sidebar":"tutorialSidebar"},"policies/ams_privacy_policy":{"id":"policies/ams_privacy_policy","title":"AMS Privacy Policy","description":"Controller details","sidebar":"tutorialSidebar"},"policies/terms":{"id":"policies/terms","title":"Terms of Use","description":"By registering as a user you declare that you have read, understood and will abide by the following conditions of use:","sidebar":"tutorialSidebar"},"policies/tom":{"id":"policies/tom","title":"Technical and organisational measures (TOM)","description":"This document describes the technical and organisational measures established by National Infrastructures for Research and Technology S.A. (GRNET S.A.) to meet legal and contractual requirements when processing personal data, conducting a higher level of security and protection.","sidebar":"tutorialSidebar"},"training_material/training_material":{"id":"training_material/training_material","title":"Training Material","description":"We have compiled a list of resources that will assist any newcomer","sidebar":"tutorialSidebar"},"use_cases/use_cases":{"id":"use_cases/use_cases","title":"Use cases","description":"The integration between different core services using the ARGO Messaging Service (AMS) as transport layer was one of our","sidebar":"tutorialSidebar"}}}')}}]);