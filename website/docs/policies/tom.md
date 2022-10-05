---
sidebar_position: 4
title: Technical and organisational measures (TOM)
---


This document describes the technical and organisational measures established by National Infrastructures for Research and Technology S.A. (GRNET S.A.) to meet legal and contractual requirements when processing personal data, conducting a higher level of security and protection.

The definitions in Article 32 of the GDPR apply.

## 1. Confidentiality

### 1.a. Physical Access Control 
Actions suitable for providing physical and environmental security of data centers, server room facilities, and working areas are adapted, while  precautions against environmental threats and power disruptions are also granted. Access is limited by job role and subject to authorised approval.  Technical and organisational measures that are taken into account involve security guard personnel, reception, doorbell systems, manual locking systems, and video surveillance of entrances.  Entrance to the building is only possible with keycards and keys provided to authorised employees or by visitors accompanied by employees, preventing unauthorised persons from accessing secured areas. Additionally, Information on security policy, work instructions for operational safety, and access control are provided.

### 1.b. Logical Access Control 
Operations sufficient for preventing data processing systems from being used by unauthorised persons are applied. Logical access controls are designed based on authority levels and job functions. Granting access is gained on a need-to-know and least privilege basis, where it is restricted to authorised employees responsible for the job. The use of unique IDs -identified through Active Directory- and passwords for all users is adapted, including a periodic review and revoking access when employment terminates or changes in job functions occur. Technical and organisational measures that are taken into account involve; username and password protected systems, intrusion detection facilities, use of Virtual Private Networks (VPNs) for remote access, firewalls, intrusion Detection System (IDS), user permission management, information security policy, work instruction of IT user regulations, operation security and access control.

### 1.c. Authorization Control 
Actions to ensure that those authorized to use a data processing system can only access the data subject to their access authorization -based on their rule- and that personal data cannot be read, copied, modified, or removed without authorization during processing, use and after storage. Technical operations incorporate physical deletion of data carriers, logging of accesses to applications, specifically when entering, changing, and deleting data, SSH encrypted access, and certified SSL encryption. At the organisational level, a minimum number of administrators is applied, management of user rights are controlled by administrators, work instruction communication security and handling of information and values are also claimed.

### 1.d. Separation Control 
It is ensured that personal data collected for different purposes can be processed separately. Multi-tenancy of relevant applications is performed, or systems are physically or logically separated. The development sheet is separated for each product, and the services have their own line of environments. All environments, documents and other data are shared for the members of that project/product, while operational, information and data protection securities and policies are applied.

### 1.e. Pseudonymization 
Operations for pseudonymization or anonymization of personal data are implemented to the extent necessary. Internal instruction to pseudonymize or anonymize personal data as far as possible in the event of disclosure or even after the statutory deletion period has expired. Specific internal regulations on cryptography, while operational, information and data protection securities and policies are applied.


## 2. Integrity

### 2.a. Transfer Control 
Measures are taken into account to ensure that personal data cannot be read, copied, altered or removed by unauthorized persons during electronic transmission or while being transported or stored on data media. As technical and organizational actions are the use of Virtual Private Networks (VPNs) and firewalls, the provision via encrypted connections and techniques such as SSH, SFTP, HTTPS and secure cloudstores, the logging of accesses and retrievals, while operational, information and data protection securities and policies are also applied.

### 2.b. Input Control 
Operations that ensure that it is possible to check and establish retrospectively whether and by whom personal data has been entered into, modified or removed from data processing systems are implemented to the extent necessary. Input control is achieved through logging, which can take place at various levels (e.g., operating system, network, firewall, database, application). Traceability of data entry, modification and deletion through individual user names, assignment of rights to enter, change and delete data on the basis of an authorisation concept, while information security policy and work instruction of IT user regulations are engaged.


## 3. Availability and Resilience

### 3.a. Availability Control 
Actions to ensure that personal data is protected against accidental destruction or loss have been implemented to the required extent. Measures comprise fire and smoke detection systems, fire extinguishers, air-conditioning, temperature and humidity monitoring and video surveillance in server rooms, UPS system and emergency diesel generators deployment, RAID system and hard disk mirroring for data backup, information security policy and work instruction operational security.

### 3.b. Recoverability Control 
Data backups of databases and operating system images are taken to the extent required and with the aim of preventing the loss of personal data in the event of physical or technical incident. Backups are performed for network drives and servers in productive operation, where the process is being recorded (logged). The backup concept is applied according to criticality and customer specifications. When applicable storage of backup media obtained in a safe place outside the server room. Information security policy and work instruction of IT user regulations are also engaged.


## 4. Procedures for regular Review, Assessment and Evaluation

### 4.a. Data Protection Management 
Technical and organizational measures that are taken into account are; central documentation of all data protection regulations with access for employees, security certification according to ISO 27001. Updates and reviews of the effectiveness of the TOMs are carried out periodically. Data protection checkpoints are consistently implemented, while data processing systems (IT systems) are checked regularly to the extent required and after changes to ensure that they are functioning properly. A Data Protection Officer (DPO) group is appointed and notified of physical or technical incidents, and staff is trained and obliged to confidentiality and data secrecy. Data Protection Impact Assessment (DPIA) is carried out as required, whereas processes regarding information obligations according to Art. 13 and 14 GDPR are established.

### 4.b Incident Response Management 
Technical and organizational actions have been established to the extent required for security breach response and data breach process. The use and the regular updating of firewall, spam filters, virus scanning, Intrusion Detection System (IDS), and Intrusion Prevention System (IPS) for customer systems on order, are served. The process for detecting and reporting security incidents and data breaches is being documented via ticket system, with regard to reporting obligation to the supervisory authority. Formalized procedure for handling security incidents, including the involvement of DPO and ISO in security incidents and data breaches is available, while operational, information, data protection and IT user regulations, securities and policies are also applied. 

### 4.c Data Protection by Design and by Default
Measures pursuant to Art 25 GDPR comply with the principles of data protection by design and by default. No more personal data is collected than is necessary for the respective purpose. Data Protection Policy (includes principles "privacy by design and by default").

### 4.d Order Control (outsourcing, subcontractors, and order processing) 
Actions to ensure that personal data processed on behalf of the client can only be processed in accordance with the client's instructions. Technical and organizational measures have been established to the required extent. Measures involve monitoring of remote access by external parties, in the context of remote support and work instruction supplier management and supplier evaluation. Moreover, a prior review of the security measures taken by the contractor and their documentation is applied. Selection of the contractor under due diligence aspects (especially with regard to data protection and data security) is achieved. Conclusion of the necessary data processing agreement on commissioned processing or EU standard contractual clauses and a framework agreement on contractual data processing within the group of companies, where written instructions to the contractor and obligation of the contractor's employees to maintain data secrecy. Additionally, an agreement on effective control rights over the contractor and regulations on the use of further subcontractors is maintained, ensuring also the destruction of data after termination of the contract or in the case of longer collaboration, ongoing review of the contractor and its level of protection.


## 5. Organization and Data Protection
The National Infrastructures for Research and Technology S.A. (GRNET S.A.), based on Its Quality and Information Security Policies has set itself the goal of providing products and services to be delivered at the highest possible level of information security in compliance with the law.  In this context GRNET S.A. has established, the roles of Information Security Officer (ISO), Data Protection Officer (DPO), Quality Officer (QO), and Legal Compliance Officer (LCO) as well as a Corporate Binding Rules (a set of internal guidelines and regulations) on information security and data protection, that are contractually binding for all employees, that defines secure information and data handling formed in secrecy and confidentiality. Employees are continuously informed and trained in the area of data protection, while third parties who may come into contact with personal data in the course of their work for GRNET S.A. are obligated to comply with data protection and data secrecy by means of a so-called NDA (Non-Disclosure Agreement) before they begin their work. Any subcontractors entrusted with further processing are only used after approval by the clients and after the conclusion of a Data Processing Agreement (DPA) in accordance with Art 28 GDPR, with which they are fully bound by all data protection obligations to which GRNET S.A. itself is subject. Current high technical security standards at GRNET S.A. are periodically reviewed and confirmed for adequacy and effectiveness in the course of ongoing internal audits and annually by Independent, External, Accredited Certification Bodies.


## 6. Certifications 
The Quality Management System (QMS) as well as the Information Security Management System (ISMS) of GRNET S.A. are both certified by Independent Accredited Certification Bodies according to ISO 9001 and ISO 27001.
