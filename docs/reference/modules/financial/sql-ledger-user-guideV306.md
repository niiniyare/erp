SQL-Ledger User Guide

#### Version 3.0.

#### First Edition

#### Written by

#### Armaghan Saqib

#### Sebastian Weitmann

#### International SQL-Ledger Network Association

#### Grundstrasse 16b, 8712 Stäfa,

#### Switzerland

#### January 1, 2015


## Contents







- 1 Introduction
   - 1.1 Introducing SQL-Ledger
      - 1.1.1 Features
      - 1.1.2 Versions
      - 1.1.3 Website and other resources on Internet
   - 1.2 Getting up and running
      - 1.2.1 Hosted version
      - 1.2.2 Quick installation script
      - 1.2.3 Manual install
      - 1.2.4 github branch structure
      - 1.2.5 Upgrading from previous versions of SQL-Ledger
      - 1.2.6 Data backup
      - 1.2.7 Data restore
   - 1.3 Our enhancements to standard SQL-Ledger
      - 1.3.1 General
      - 1.3.2 Departments
      - 1.3.3 Warehouses
      - 1.3.4 COGS
      - 1.3.5 Reports
   - 1.4 Future development of SQL-Ledger
      - 1.4.1 Bug fixes
      - 1.4.2 New features
      - 1.4.3 New version
- 2 Setting up your business on SQL-Ledger
   - 2.1 Creating your first dataset
   - 2.2 Creating users and roles
      - 2.2.1 Roles
         -
- CONTENTS
      - 2.2.2 Users / Employees
   - 2.3 Defaults
   - 2.4 Customers
      - 2.4.1 Adding a new customer
      - 2.4.2 Editing or deleting an existing customer
   - 2.5 Vendors
      - 2.5.1 Adding a new vendor
      - 2.5.2 Editing or deleting an existing vendor
   - 2.6 Type of Business
   - 2.7 Departments
      - 2.7.1 Managing Departments
      - 2.7.2 Default Department
      - 2.7.3 Using Departments
      - 2.7.4 Using Departments in Reports
   - 2.8 Projects
      - 2.8.1 Managing Projects
      - 2.8.2 Using Projects
      - 2.8.3 Project Reports
   - 2.9 Chart of Accounts
      - 2.9.1 Heading accounts
      - 2.9.2 Account types
      - 2.9.3 Marking accounts
      - 2.9.4 Mandatory default accounts
      - 2.9.5 GIFI
   - 2.10 Templates
      - 2.10.1 Editing Templates
      - 2.10.2 Template Variables
      - 2.10.3 Template control commands
      - 2.10.4 Three type of templates:
      - 2.10.5 An Introduction to LATEX
      - 2.10.6 Structure of a LATEX Document
   - 2.11 Goods & Services
      - 2.11.1 Parts
      - 2.11.2 Services
      - 2.11.3 Labor/Overhead
      - 2.11.4 Assemblies
      - 2.11.5 Groups
      - 2.11.6 Pricegroups
- CONTENTS
   - 2.12 Goods & Services Reports
      - 2.12.1 All Items
      - 2.12.2 Parts
      - 2.12.3 Requirements
      - 2.12.4 Services
      - 2.12.5 Labor/Overhead
      - 2.12.6 Groups
      - 2.12.7 Pricegroups
      - 2.12.8 Assemblies
      - 2.12.9 Components
      - 2.12.10Stock Assembly
   - 2.13 Warehouses
      - 2.13.1 Adding warehouses
      - 2.13.2 Default warehouse
      - 2.13.3 Using warehouses
      - 2.13.4 Warehouse transfers
      - 2.13.5 Transfer Reports
      - 2.13.6 Warehouse Onhand Report
      - 2.13.7 Activity Report
   - 2.14 Languages
   - 2.15 Translations
   - 2.16 Taxes
      - 2.16.1 Define the tax accounts in chart of accounts
      - 2.16.2 Define tax percentages
      - 2.16.3 Mark Items/Services as taxable
      - 2.16.4 Mark Customers/Vendors for applicable taxes
   - 2.17 Data import from other applications
      - 2.17.1 Sale invoices
      - 2.17.2 Receipts and Payments
      - 2.17.3 AR/AP Transactions
      - 2.17.4 General Ledger
      - 2.17.5 Customers and Vendors
      - 2.17.6 Parts
      - 2.17.7 Vendor price list
      - 2.17.8 Customer price list
      - 2.17.9 Chart of accounts
- CONTENTS
- 3 Running your business on SQL-Ledger
   - 3.1 AR
      - 3.1.1 AR Transaction
      - 3.1.2 Sales Invoice
      - 3.1.3 Credit invoice and credit note
   - 3.2 AR reports
      - 3.2.1 Transactions report
      - 3.2.2 Aging report
      - 3.2.3 Reminders
      - 3.2.4 Customer history reports
   - 3.3 Point of sales (POS)
      - 3.3.1 Creating a POS invoice
      - 3.3.2 Viewing open invoices
      - 3.3.3 Receipts
   - 3.4 AP
      - 3.4.1 AP transactions
      - 3.4.2 Vendor invoice
      - 3.4.3 Debit invoice and debit note
   - 3.5 AP reports
      - 3.5.1 Transactions report
      - 3.5.2 Aging report
      - 3.5.3 Vendor history reports
   - 3.6 Cash
      - 3.6.1 Receipts
      - 3.6.2 Payments
   - 3.7 General ledger
      - 3.7.1 Add transaction
      - 3.7.2 Reports
   - 3.8 Recurring transactions
      - 3.8.1 Scheduling
      - 3.8.2 Generating
   - 3.9 Currencies and exchange rates
      - 3.9.1 Defining currencies
      - 3.9.2 Buying and selling in foreign currencies
      - 3.9.3 Reports
      - 3.9.4 Exchange rate difference
      - 3.9.5 Fund transfers in foreign currencies
   - 3.10 Quotations and RFQs
- CONTENTS
      - 3.10.1 Quotations
      - 3.10.2 RFQ
   - 3.11 Orders
      - 3.11.1 Sales orders
      - 3.11.2 Purchase orders
      - 3.11.3 Important inventory on-hand reports
   - 3.12 Time Cards
      - 3.12.1 Create a project for the customer
      - 3.12.2 Create time card entries
      - 3.12.3 Create a sales order for open time cards
   - 3.13 Audit Control
      - 3.13.1 Enforce transaction reversal for all dates
      - 3.13.2 Close books up to
      - 3.13.3 Activate audit trail
      - 3.13.4 Remove audit trail up to
   - 3.14 Reconciliation
      - 3.14.1 Marking transactions
   - 3.15 Year end
   - 3.16 Data backup
      - 3.16.1 Send by Email
      - 3.16.2 Save to File
   - 3.17 Basics of double entry accounting
      - 3.17.1 Introduction
      - 3.17.2 Account types
      - 3.17.3 Accounting rules
      - 3.17.4 Examples
   - 3.18 Cost of goods sold (COGS)
      - 3.18.1 Sale invoices and COGS
      - 3.18.2 Sales before purchases
      - 3.18.3 Editing sale invoices
   - 3.19 Ledger Doctor
   - 3.20 Monitor
- 4 Keeping track of your business in SQL-Ledger
   - 4.1 Financial reports
      - 4.1.1 Chart of accounts & trial balance
      - 4.1.2 Income statement
      - 4.1.3 Balance sheet
- CONTENTS
      - 4.1.4 Tax report
      - 4.1.5 Project & department income statement
      - 4.1.6 Project Income statement
      - 4.1.7 Department Income statement
   - 4.2 Module reports
      - 4.2.1 AR reports
      - 4.2.2 Customers reports
      - 4.2.3 AP reports
      - 4.2.4 Vendor reports
      - 4.2.5 Cash reports
      - 4.2.6 Order entry reports
      - 4.2.7 Warehouses reports
      - 4.2.8 Quotations reports
      - 4.2.9 General ledger reports
      - 4.2.10 Project reports
- 5 Ledger Cart
   - 5.1 Introduction
      - 5.1.1 Features
      - 5.1.2 Limitations
      - 5.1.3 Using LedgerCart as an online store
      - 5.1.4 Using LedgerCart as Self service portal
      - 5.1.5 Screen shots
   - 5.2 Installation
      - 5.2.1 Software packages
      - 5.2.2 Configuration and Admin access
      - 5.2.3 Customization
- 6 Development and Customization
   - 6.1 Customization
      - 6.1.1 custom_xx.pl files
      - 6.1.2 Modify the source code
   - 6.2 Adding a new translation
   - 6.3 SQL Queries
      - 6.3.1 Simple SQL Queries
      - 6.3.2 Advanced SQL Queries
      - 6.3.3 Queries to troubleshoot database problems
   - 6.4 API
- CONTENTS
   - 6.4.1 Introduction
   - 6.4.2 API Uses
   - 6.4.3 Calling from PHP


# About authors

## Sebastian Weitmann

Sebastian Weitmann has been associated with SQL-Ledger since 2004. He has
a degree in law and has specialized in finance and accounting. He has been
implementing SQL-Ledger for businesses which range from small companies to
large corporations using SQL-Ledger to its full potential with all the features.
Sebastian strongly believes in the merits of free software for businesses and for
the society in general. In 2010 he and Thomas Brändle founded the ’International
SQL-Ledger Network Association’ (ISNA), a non-profit organisation to fund and
support the further development of SQL-Ledger.
Sebastian lives in Cologne, Germany and can be reached through his email
sw@sql-ledger.net.

## Armaghan Saqib

Armaghan Saqib is using SQL-Ledger since 2004. He has a degree in mathe-
matics and is a self taught computer programmer. In 2004 he was looking for
an open source accounting solution and he discovered SQL-Ledger. The sim-
plicity of SQL-Ledger while still being feature rich made him a fan of this ERP
solution. Since then he has written lots of code adding many new modules and
enhancements to the stock SQL-Ledger.
He is lead developer and maintainer of his SQL-Ledger version (ledger123)
which is the official SQL-Ledger fork of the International SQL-Ledger Network
Association. His consulting company Ledger123.com provides development and
implementation services to clients all over the world.
Armaghan lives in Lahore, Pakistan and can be reached through his email
saqib@ledger123.com.

##### 8


# Corporate sponsors

We would like to thank **Run my Accounts AG** from Switzerland for their
work in the past couple of years in contributing modules, improvements and bug
fixes for SQL-Ledger. Run my Accounts provides accounting services for Swiss
entities on the base of SQL-Ledger. They have developed a highly automated
accounting engine, which processes documents and bank feeds very efficiently.
Run my Accounts AG is growing at a fast pace in the Swiss market and probably
holds the largest SQL-Ledger customer base. For more information please visit
[http://www.runmyaccounts.com.](http://www.runmyaccounts.com.)

##### 9


# SQL-Ledger User Guide

Written by Sebastian Weitmann and Armaghan Saqib
Copyright © 2015 Sebastian Weitmann and Armaghan Saqib. All rights
reserved.
Published by International SQL-Ledger Network Association, Grundstrasse
16b, 8712 Stäfa, Switzerland.
While every precaution has been taken in the preparation of this book, the
publisher and authors assume no responsibility for errors or omissions, or for
damages resulting from the use of the information contained herein.

##### 10


# Preface

This manual describes how to install and use SQL-Ledger. It gives an overview
of the various features that are available in SQL-Ledger and explains how to use
them.
We wrote this manual to support the free distribution of SQL-Ledger and to
help existing SQL-Ledger users broaden their knowledge base. We have divided
it into 6 chapters.
The first chapter introduces SQL-Ledger and explains how to get it up and
running. The second chapter tells you how to set up and adapt SQL-Ledger to
fit your own business needs. It explains how to create users, customers, vendors
and everything else you need to do before you start working.
The third chapter highlights how to process your day to day business trans-
actions in SQL-Ledger. It will tell you how to register invoices, how to make a
general ledger entry and how to enter all your other related business transactions.
The fourth chapter explaines how to take advantage of all the information
you have entered, in other words, how to keep track of your business. You will
learn how to start analyzing your business data and transactions that are stored
in the database.
The fifth chapter provides you with information about LedgerCart. Ledger-
Cart is a very nice add-on tool that can instantly upgrade your SQL-Ledger
installation into a fully functional web-based ordering system and customer por-
tal.
The sixth and last chapter gives you an introduction to SQL-Ledger develop-
ment and customization.

##### 11


# Chapter 1

# Introduction

### 1.1 Introducing SQL-Ledger

SQL-Ledger is a free ERP (Enterprise Resource Planning) double entry account-
ing software with a rich set of business intelligence features. It supports multiple
users, multiple languages, multiple currencies, multiple companies, accounts re-
ceivable, accounts payable and stock tracking.
It’s a web-server application that has already been translated into 45 lan-
guages and enables business management and administration over the Internet
or on a private local network. Any web browser can be used as its standard user
interface which makes SQL-Ledger platform independent and usable on practi-
cally any operating system.
SQL-Ledger is written in the programming language Perl, runs on any modern
web server like Apache and stores all business data in a PostgreSQL database. Its
version 1.0 was released in Jan. 29, 1999. So as of this writing in 2015, it is 16
years old software which has been under constant development and enhancement
during this period. This makes it suitable enough for small as well as for large
businesses.
SQL-Ledger is licensed under the GNU GENERAL PUBLIC LICENSE com-
monly known as the GPL.

#### 1.1.1 Features

SQL-Ledger has an impressive feature set which even many commercial / propri-
etary ERP solutions don’t provide. Its internal design and user interface is quite

##### 12


##### CHAPTER 1. INTRODUCTION 13

simple which make it easy to learn. Some of these features are listed below:

1. Browse-based interface. Accessible from any browser on desktop PCs or
    mobile devices.
2. Multi-user and multi-company with a single installation.
3. Fine grained access control for each user. Users see only features/functions
    which they are allowed to by the admin.
4. Invoices with packing lists and pick lists. Invoices can be emailed in pdf or
    html formats.
5. Cash management with comprehensive bank reconciliation.
6. Powerful accounts receivables and accounts payable function with aging,
    outstanding reports, statements as well as multiple reminders for cus-
    tomers.
7. Orders management integrated with invoices with partial or full delivery /
    shipping function.
8. Quotations and Request for Quotations which are fully integrated with
    orders and invoices.
9. Departmental accounting. You can restrict users to a specific department
    so that they can add or view transactions in their respective departments
    only. Financial reports can be displayed for specific department or all
    departments.
10. Projects and time cards to track time and generate sales orders.
11. Vouchers with approval work flow. Vouchers allow you to add payables,
payments and general ledger transactions which need to be approved before
posting.
12. Inventory / stock management with invoices/orders/shipping/transfers at
multiple warehouses.
13. Assemblies with components and sub-assemblies. Suitable for manufactur-
ing businesses.


##### CHAPTER 1. INTRODUCTION 14

14. Import data for items, accounts, customers, vendors, invoices and trans-
    actions from your old accounting software.
15. Template driven documents (invoices, orders, quotes, financial reports).
    Documents can be sent to printer, file or fax, emailed or displayed on
    screen. Templates can be created in html, xml, tex, and text format.
16. Multiple currencies with automatic calculations of gain/loss on exchange
    rate differences.
17. Multi language user interface and templates.
18. Multi language descriptions for items, accounts so that your foreign cus-
    tomers get invoices in their native language.
19. User customizable reports with selection of columns and data filter before
    running the report.
20. Instant posting of COGS so that you can immediately see the profitability
    of your business through financial reports.
21. Financial reports (balance sheet, income statement) with single or multiple
    months or with comparison with past months or years.

... and much more.

#### 1.1.2 Versions

The current release of stock SQL-Ledger developed by DWS Systems Inc. is
3.0.6. The International SQL-Ledger Network Association (ISNA) also maintains
its own version of SQL-Ledger. It’s developed openly by ISNA on github.com
and is based upon the original SQL-Ledger version 3.0.6 with its enhancements.
We call it the SQL-Ledger Network version, or Ledger123 release 3. The SQL-
Ledger Network version alias Ledger123 tries to incorporate all the goodness
which comes from stock SQL-Ledger. So you get best of both worlds.
To make things simple, we assume that you are using the SQL-Ledger Net-
work version (Ledger123 release 3). Though most of the sections would apply
equally well to the stock SQL-Ledger 3.0.6 as well as older versions. This is
particularly true if you are not using inventory related functions, because most of
the enhancements in the SQL-Ledger Network version are related to inventory.


##### CHAPTER 1. INTRODUCTION 15

#### 1.1.3 Website and other resources on Internet

International SQL-Ledger Network Association website

- [http://www.sql-ledger-network.com/](http://www.sql-ledger-network.com/)

International SQL-Ledger Network Association Support Forum

- [http://forum.sql-ledger-network.com/](http://forum.sql-ledger-network.com/)

Mailing list

- [http://www.ledger123.com/mailing-list/](http://www.ledger123.com/mailing-list/)

Github repository for the SQL-Ledger Network version (Ledger123 release 3)

- https://github.com/ledger123/ledger123/tree/rel

Ledger123 Website

- [http://www.ledger123.com/](http://www.ledger123.com/)

DWS Systems Inc. SQL-Ledger website

- [http://www.sql-ledger.com/](http://www.sql-ledger.com/)

### 1.2 Getting up and running

There are three ways to get up and running with SQL-Ledger:

1. Hosted version.
2. Quick installation script.
3. Manual install.

Before you install SQL-Ledger on your chosen operating system, you first need to
install a set of other software applications that work together with SQL-Ledger
and create the base for it to function. SQL-Ledger requires the following:

- Web Server (Apache, NCSA, httpi, thttpd, ... );
- Perl (version 5 or newer);


##### CHAPTER 1. INTRODUCTION 16

- Database Server (PostgreSQL version 7.1 or newer)
- Database Driver (DBD-Pg)
- Database Independent Interface (DBI)
- LaTex (optional)

You will find various SQL-Ledger step-by-step installation guides on the Interna-
tional SQL-Ledger Network website (www.sql-ledger-network.com).

#### 1.2.1 Hosted version

Using the free hosted SQL-Ledger instance, available from ISNA website is the
easiest way to get up and running with SQL-Ledger. To get this, visit sql-ledger-
network.com and signup for your free instance. We shall create your free account
and will send you user name and password. You can then login and start adding
transactions.

#### 1.2.2 Quick installation script

We have quick installation bash scripts for many Linux distrubitions which will
quickly install SQL-Ledger on a freshly installed Linux desktop or server. Visit
sql-ledger-network.com and download the appropriate script for your Linux dis-
tribution. Once you have download the script you can run it in the terminal
window using the following command:

# bash install_ledger

As mentioned above the only requirement for this method is a freshly in-
stalled supported Linux distribution on a PC. Once installed you can access your
installation on ’http://yourserver_ip/ledger123’. where ’yourserver_ip’ is the ip
address of your server. You can substitute ’localhost’ if you are using your server
machine as your desktop too.

#### 1.2.3 Manual install

In section we shall show how you can install SQL-Ledger on your existing server
or desktop by installing and configuring all packages and services yourself.


##### CHAPTER 1. INTRODUCTION 17

**1.2.3.1 Installing the SQL-Ledger Network version using ’git clone’**

The recommended way to download and install the SQL-Ledger Network version
is to use the ’git’ package. To install git on Ubuntu, you run ’sudo apt-get install
git-core’. Once git is successfully installed, you can follow these steps:

1. Download the SQL-Ledger Network GitHub repository. You will get a fully
    working SQL-Ledger installation which includes our enhancements. (The
    default ’master’ branch)

```
git clone git://github.com/ledger123/ledger123.git
```
2. You are probably interested in the latest release 3. The following command
    will switch to it.

```
git checkout -b rel3 origin/rel
```
3. From now onwards you can upgrade to our latest enhancements, as well
    as new features released by sql-ledger.com, with the following simple com-
    mand:

```
git pull
```
4. Let us say you are not interested in our enhancements and just want to
    maintain and upgrade to the SQL-Ledger release from SQL-Ledger.com.
    Switch to the SQL-Ledger branch first time:

```
git checkout -b sql-ledger origin/sql-ledger
```
5. You can also switch back to any past SQL-Ledger version. First see a log
    of all commits and 40 chars hashes:

```
git log --pretty=oneline
```
6. To revert to SQL-Ledger 2.8.17 type


##### CHAPTER 1. INTRODUCTION 18

```
git checkout 7b15e9b
```
Note: To view all code changes, you can visit https://github.com/ledger123/ledger
and select the chosen branch.

**1.2.3.2 Installing Perl modules**

Future versions of our enhanced SQL-Ledger release may add dependencies to
some Perl modules. Before adding any such dependencies we shall make sure
these Perl modules can be installed on most common Linux distributions without
much hassle.
There are three ways to install any Perl module:

1. Install the pre-built module package for your Linux distributiion package
    manager (apt-get or yum)

```
apt-get install libdbix-simple-perl # ubuntu / debian
```
2. Install using cpan command. cpan command comes built-in with the Perl
    installation on all distributions.

```
cpan DBIx::Simple
```
```
You may need to answer to few configuration questions when you are
running cpan for the first time.
```
3. Install using cpanm (cpanminus) which is relatively less complicated than
    cpan. You can install cpanm with following command:

```
curl -L http://cpanmin.us | perl - App::cpanminus
```
```
Once installed, you can install Perl module of your choice using cpanm
followed by the module name like:
```
```
cpanm DBIx::Simple
```

##### CHAPTER 1. INTRODUCTION 19

#### 1.2.4 github branch structure

Our main repository is available at https://github.com/ledger123/ledger123/.
Following branches are of interest to you. (The other branches on this repository
are either obsolete or contain some custom code so can be ignored without
worrying about them.)
Branch Description
master This is version 2.8 branch which is rarely updated now. You should
upgrade to version 3 which is branch ’rel3’.
sql-ledger This branch contains unmodified code from Dieter’s SQL-Ledger from
version 2.6.0 to the current version released on sql-ledger.com.
rel3 This branch is the latest enhancd version of SQL-Ledger which we
recommend to use. This version contains all the updates by Dieter’s
SQL-Ledger as well as all our enhancements. This manual assumes that
you are using this version.

#### 1.2.5 Upgrading from previous versions of SQL-Ledger

You need to take following two steps to upgrade to the latest enhanced SQL-
Ledger (ledger123).

1. Upgrade your installation to the latest release ’rel3’ using git as described
    in 1.2.3.
2. Add the database changes for enhanced version using the following com-
    mand:

```
psql -U sql-ledger your-dataset-name < sql/Pg-custom_tables.sql
```
Notes:

1. The above statement assumes that you have a PostgreSQL user named
    ’sql-ledger’ for use with your SQL-Ledger datasets.
2. As a precaution you should always backup your data before any upgrade.
    See 1.2.6 for details on how to do this.


##### CHAPTER 1. INTRODUCTION 20

#### 1.2.6 Data backup

SQL-Ledger stores all its data in three places:

1. All business data is stored in a PostgreSQL database. You can backup
    this data using the built-in backup menu (see 3.16) or using the pg_dump
    command. This backup is the critical part of running your SQL-Ledger and
    should be done every day or every few hours depending upon the usage of
    your SQL-Ledger installation. It is recommended to setup a cron job to do
    this backup automatically at the scheduled time of the day and copy the
    backup file to some other computer or storage media.
    You can use the following command to backup your PostgreSQL data.
    pg_dump -U sql-ledger dataset-name > dataset-name.sql

```
or better yet with gzip compression:
```
```
pg_dump -U sql-ledger dataset-name | gzip -c > dataset-name.sql.gz
```
2. User information (preferences and passwords) is stored in a text file ’user-
    s/members’. As part of your backup, you can copy this file to a safe place.
    This file can be re-created by opening each user information using ’ **HR–**
    **Employees** ’ menu and saving again with a new password but it can be
    long and tedious process and cause disruption to the usage of SQL-Ledger
    so it is always better to have the latest copy of this text file in a safe place.
3. Templates (html, latex and text) specific to your company dataset are
    stored in ’templates/DATASET-NAME’ folder. You should backup this
    folder regularly whenever you make changes to your templates. If you
    don’t backup this folder then you will need to recreate these templates
    which can be a long and tedious process.
    You can use the following command to backup your users and templates
    data.
    tar czvf backup.tar.gz users/members templates/dataset-name/

```
Tip: There is no harm in having complete backup of your SQL-Ledger installation
folder. This will ensure that you get all templates and users as well as the correct
version of SQL-Ledger when you are trying to restore the backup in the time of
need. You will, however, still need to backup your database separately as shown
above.
```

##### CHAPTER 1. INTRODUCTION 21

#### 1.2.7 Data restore

There are two possible scenarios when you need to restore your data.

1. You have your SQL-Ledger running smoothly and just want to restore some
    old backup of the same dataset.
    For this you will first delete your existing data base (make sure you have
    latest backup) and then recreate database and restore old backup into it
    using commands similar to the following:

```
dropdb -U sql-ledger dataset-name
createdb -U sql-ledger dataset-name
psql -U sql-ledger dataset-name < your-dataset-backup-file.sql
```
2. You have setup a new server and want to restore SQL-Ledger with data,
    users and templates.
    To proceed with this you will install SQL-Ledger again (or just untar/unzip
    the existing sql-ledger folder as mentioned in the backup section tip), re-
    store your database and then copy users/members file and templates folder
    in the appropriate place. You can using commands similar to the following
    to achieve this.

```
createdb -U postgres sql-ledger dataset-name
psql -U sql-ledger dataset-name < dataset-name.sql
# Install sql-ledger again or untar/unzip sql-ledger folder
# as mentioned in the tip above.
cd sql-ledger; tar xzvf backup.tar.gz
```
You may need to make some adjustments to the commands shown above.

### 1.3 Our enhancements to standard SQL-Ledger

#### 1.3.1 General

1. A new pleasant looking css theme.
2. jQuery based lookups of items, customers, vendors without requiring a
    screen update to select these items.


##### CHAPTER 1. INTRODUCTION 22

3. Calendar for date lookup.
4. ’Add Customer’, ’Add Vendor’ links on invoices/orders/quotes/POS screens.
5. Enhanced assemblies: You can get a report of all stock-assembly actions.
    Inventory at warehouses is correctly updated with any assemblies made and
    components used.
6. Enhanced bank reconciliation.
7. Added back the ’ **Shipping–Transfer** ’ function from SQL-Ledger 2.6.
8. LedgerDoctor script which identifies potential problems with data entry.
9. CSV data import with download-able sample data files. (invoices, trans-
    actions, general ledger, orders, customers, vendors, parts, chart).
10. Disabled incorrect item weight update from orders and invoices.
11. Parts group is mandatory if there is at least on group defined.

#### 1.3.2 Departments

1. Restrict user to a particular department using ’ **HR–Employees** ’ menu.
2. Default department for user.
3. Department is mandatory on invoices/orders/quotes if there is at least one
    department defined.

#### 1.3.3 Warehouses

1. Warehouse transfers module.
2. Restrict user to a particular warehouse ’ **HR–Employees** ’ menu.
3. Default warehouse for user.
4. Track warehouse inventory from sales and purchase invoices.
5. Track inventory-in-transit between warehouse movement.


##### CHAPTER 1. INTRODUCTION 23

6. Warehouse is mandatory on invoices if there is at least one warehouse
    defined.
7. Warehouse on-hand and activity reports.

#### 1.3.4 COGS

1. Re-posting script which corrects cogs errors due to invoice editing or post-
    ing an invoice in the past date.
2. Per Invoice and per invoice item cogs/revenue information with gross profit
    %age.
3. On-hand value report which shows the inventory onhand quantities and
    value based upon FIFO costing.

#### 1.3.5 Reports

1. Per-invoice and per-item cogs/revenue information.
2. Enhanced tax reports.
3. Audit trail report.
4. Drill-down to transactions from income statement.
5. Invoice date and customer/vendor filter in ‘All Items’ report.
6. Account description in ‘GL Reports’.
7. Account activity report using ‘GL Reports’.

### 1.4 Future development of SQL-Ledger

#### 1.4.1 Bug fixes

Bug fixes are done always immediately when we discover a bug ourselves or it
has been reported by a user. You can report bugs on our github issues tracker
at https://github.com/ledger123/ledger123/issues. Add the issue and label it as
’bug’.


##### CHAPTER 1. INTRODUCTION 24

#### 1.4.2 New features

New features are being added to the release 3 on regular basis. You can add the
required feature on our github issues tracker at https://github.com/ledger123/ledger123/issues.
Add your request there and label it as ’enhancement’. For immediate availability
of your preferred feature you can sponsor its development.

#### 1.4.3 New version

We are in the planning stages of an entirely new SQL-Ledger from ground up
using a modern web application development framework. In this respect we have
decided to:

1. Using Mojolicious as our framework of choice for the development of new
    version.
2. Be 100% compatible with the release 3 (the current version) during the
    whole development process so that there is virtually no upgrade path re-
    quired to switch to the new version at any time. You should be able
    to switch to either new version or old version without compromising any
    feature or losing any data.

Details will be shared on our mailing list and forum as we make progress with
this work.


# Chapter 2

# Setting up your business on

# SQL-Ledger

The next step after successful SQL-Ledger installation is to setup your initial
business data. You need to do this before you can start making your day to day
transactions.

### 2.1 Creating your first dataset

```
If you are using free hosted version on ISNA website, you can skip this section
and go directly to 2.2.
```
Following instructions assume that you have installed SQL-Ledger using man-
ual method (see 1.2.3). You need to create a dataset in SQL-Ledger before you
can start using it to manage your business accounts. Behind the scenes each
dataset is a PostgreSQL database with tables, indexes and some seed data like
chart of accounts.
To create a new dataset, you need to login to the admin interface. The admin
interface is accessible through the following URL:

- [http://your-server-ip-address/sql-ledger/admin.pl](http://your-server-ip-address/sql-ledger/admin.pl)

You will be asked for a password. The default password is blank. Once you login
for the first time, you are asked to set the password to something secure.

##### 25


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 26

Once you have logged-in to the admin interface, you will see the existing
datasets if any.

To add your new dataset, you click ’Add Dataset’ button and following page
is displayed.

On this page you need to enter your database credentials. As a best practice
you should create a PostgreSQL user for use with SQL-Ledger. By default this
user is assumed to be ’sql-ledger’. In Debian or Ubuntu distributions you can
create this user by entering the following the command:


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 27

su postgres -c "createuser -d -S -R sql-ledger"

If you are not sure how to create this user on your own system, you can
go ahead with the default PostgreSQL superuser, which is normally ’postgres’,
instead of ’sql-ledger’ and click ’Continue’ button. All the other defaults on the
above page will work in most cases.

Once the above page is displayed, you can enter your company name, a name
for your dataset (which should be in lowercase without any spaces or special
characters), the character encoding and one of the default chart of accounts.
Once you have made all the selections, click ’Continue’ to create your dataset.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 28

```
Your dataset will be created and added to the list of existing datasets.
```
### 2.2 Creating users and roles

A default admin account with name ’admin@datasetname’ is created with each
new dataset. Its password is set to blank so make sure to change it to something
secure on your first login. Now you need to login and create some new users as
well as set up their access privileges using roles.
To login to your newly created dataset visit [http://your-server-ip-address/sql-](http://your-server-ip-address/sql-)
ledger/login.pl and login as ’admin@datasetname’ without specifying any pass-
word.
Tip: At this point you can start using SQL-Ledger for your business by creating
invoices and other transactions but it is better to create a ’regular’ user and use
it instead of using admin user for this purpose. Admin user should be reserved
for performing administrative tasks only.

#### 2.2.1 Roles

Roles allow you to define which menus are available to each user. You can group
your users into roles and then define the access privileges for the roles. Click the
menu **’System–Roles’** to display existing roles, change them or add a new one.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 29

To add a new role, click ’Add Role’ and check/un-check the menus under
’Access Control’ to allow or disallow that menu to the role. If you un-check
for example ’AR’, all features pertaining to ’AR’, like ’Add Transaction’ , ’Sales.
Invoice’ etc. will also be disabled. Once you have defined the access privileges,
click ’Save’ to add the role.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 30

#### 2.2.2 Users / Employees

Once you have defined the roles, it is time to define the actual users. For this
you click on **’HR–Employees–Add Employee** ’. Here you fill all the information
for the user.

- In the ’Role’ field, select the appropriate role for this user. If no role has
    been created then user will have access to all the menus.
- In the login name field, type the login name which should be preferably in
    lower case without ’@’ sign and without other special characters.
- The ’Sales’ check-box is there to mark whether this user is to appear as a
    salespersons on your quotations, orders and invoices or not.

On the screen you can add all your users as well as other employees. If you do
not want to allow a particular employee to login to SQL-Ledger, just omit the
login and password.
Tip: Instead of using roles you can also define menu access for each user indi-
vidually using the ’Access Control’ button on add/edit employee screen. This
is particularly useful if you have only 1 or 2 users. With large number of users,
however, it will become cumbersome to manage users access individually and use
of roles is recommended.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 31

To get list of existing employees, you use **’HR–Employees–Reports’** menu.
A search screen is displayed where you can select which information you need to
display. To display the list click the ’Continue’ button.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 32

Now all employees and/or users are displayed. You can click on the name of
the user to open the detail of the user or employee and change it as required.

Once you have created a user, he or she can login with his or her username.
The username is of the format ’login@datasetname’. For example if you have
created a user with login ’armaghan’ for a dataset named ’rel3’ then the user
needs to login with ’armaghan@rel3’ as his or her username.
If you use SQL-Ledger in a multi-company environment and a user has access
to various different datasets, by entering only the username without the ’@’ and
without the name of the dataset, you will get a list of all the different datasets
(companies) available to choose from.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 33

When you or your user has successfully logged-in to SQL-Ledger, the welcome
screen (shown below) is displayed. The menu is on the left. Only those menu
options are visible to the user which have been allowed by the assigned role to
that user (see 2.2 above).


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 34

### 2.3 Defaults

The **’System–Defaults’** menu allows you to setup your company name, address
and business information and defaults in SQL-Ledger. Document numbering is
also controlled by system defaults. You setup defaults for document numbers as
shown on the screen shot below. You can change these according to your business
needs. You can also use following variables in the system default number fields:

<%DATE%>
<%YYMMDD%>
<%YEAR%>
<%MONTH%>
<%DAY%>
<%NAME 1 1 3%>


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 35

##### <%BUSINESS%>

##### <%BUSINESS 10%>

##### <%CURR...%>

##### <%DESCRIPTION 1 1 3%>

##### <%ITEM 1 1 3%>

##### <%PARTSGROUP 1 1 3%>

##### <%PHONE%>

##### <%YY%>

##### <%MM%>

##### <%DD%>

##### <%FDM%>

##### <%LDM%>


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 36

### 2.4 Customers

The customers menu allows you to add new customers and change or delete
existing customers. You need to add at least one customer before creating a
quotation, sales order, sales invoice or AR transaction.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 37

#### 2.4.1 Adding a new customer

Use **’Customers–Add Customer’** to add new customers.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 38

#### 2.4.2 Editing or deleting an existing customer

To make changes to existing customers, first you list them using **’Customers–
Customers Search** ’. You can leave this search form blank and click ’Continue’
to get all customers or you can specify a customer name, phone number or any
other information to get a specific customer. If there are more than one matching


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 39

customers, all those will be listed.

In this report you can click on the customer name and its details will be
opened in a new screen where you can make changes to existing data or delete
the customer. Please note that you cannot delete a customer once you have
posted invoices or transactions to this customer.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 40

### 2.5 Vendors

The vendors menu allows you to add new vendors and change or delete existing
vendors. You need to add at least one vendor before creating a request for
quotation (RFQ), purchase order, vendor invoice or AP transaction.

#### 2.5.1 Adding a new vendor

Use **’Vendors–Add Vendor’** to add new vendors.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 41

#### 2.5.2 Editing or deleting an existing vendor

To make changes to existing vendors, first you list them using **’Vendors–Vendors
Search** ’. You can leave this search form blank and click ’Continue’ to get all
vendors or you can specify a vendor name, phone number or any other informa-
tion to get a specific customer. If there are more than one matching vendors, all


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 42

those will be listed.

In this report you can click on the vendor name and its details will be opened
in a new screen where you can make changes to existing data or delete the
vendor. Please note that you cannot delete a vendor once you have posted some
invoices or transactions to this vendor.

### 2.6 Type of Business

You can group your customers and set default discount percentages for them
using ’type of business’ codes. To create or change these codes you use **’System–**


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 43

**Type of business’** menu.

Once you have defined types of business, you can specify them for a customer
when you are adding a new customer or editing an existing one as shown below.

### 2.7 Departments

Departments are optional and can be used to classify transactions according
to a department code. When managing departments, the following points are
important:

1. Departments lookup does not appear on transaction forms as well as on


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 44

```
search screens unless you have defined at least one department.
```
2. SQL-Ledger departments can be mapped to the various departments (sales,
    purchase etc.), branches (London, Oxford etc.) or product divisions (Prod-
    uct 1, Product2 etc.) within your organization.
3. Departments can be marked as ’Cost Center’ or ’Profit Center’. Cost center
    departments appear only in purchasing modules. Profit center departments
    appear both in purchasing and sales modules.
4. You can also change ’Department’ to anything you like (eg.Branch) using
    the SQL-Ledger language customization feature. See 6.2 on how to do
    this.

#### 2.7.1 Managing Departments

You get a list of existing departments when you click on **’System–Departments** ’
menu.

On this list you can click on department name to change its description or
delete it. You can also click on the ’Add Department’ button to add a new depart-
ment. Please note that you cannot delete a department if there are transactions
which reference this department.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 45

#### 2.7.2 Default Department

You can restrict a user to a particular department when you add a new user or
modify an existing user using the **’HR–Employees’** menu.

#### 2.7.3 Using Departments

Once departments are defined you can specify them in your invoices, orders,
quotations and other transactions.

#### 2.7.4 Using Departments in Reports

Most reports allow you to view all or department specific transactions. For
example you can filter AR Transactions report ’ **AR–Reports–Transactions** ’ by
specifying a particular department on the search screen.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 46

’ **Reports–Income Statement** ’ and ’ **Reports–Balance sheet** ’ can be fil-
tered and compared by department.

The **’Reports–Department Income Statement’** report shows income state-
ment for individual departments in columnar form as shown below:


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 47

### 2.8 Projects

Projects are optional and can be used for the following things:

1. Track income and expenses to specific projects using invoices and transac-
    tions.
2. Job costing.
3. Enter time card data.

Please note that projects only appear on transaction forms and search screens if
you have created at least one project.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 48

#### 2.8.1 Managing Projects

You can add or change projects through the ’ **Projects** ’ menu. Click on **’Projects–
Add Project’** to add a new project.

To change an existing project, first you need to display a list of your existing
projects. For this you use the **’Projects–Reports–Projects’** menu and the
following screen is displayed where you can specify some conditions to select the
projects of your interest. To view all projects just click the ’Continue’ button.

Once projects are displayed as shown below, you can click on the project
number to open the project and make the required changes.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 49

#### 2.8.2 Using Projects

Once you have defined projects, you can use them in quotations, orders, invoices
and general ledger entries. In quotations, orders and invoices the project drop-
down appears on the extended detail lines. To display this extended detail line
you first need to check the check-box next to the description of each line item.
If you check the check-box in the heading as shown below and click ’Update’
button, the extended details are shown for all line items.

#### 2.8.3 Project Reports

The **’Projects–Reports–Transactions’** report will show you a summary report
similar to the ’ **Reports–Trial Balance** ’ report with summary of all transactions
for the the selected project. Before displaying the report you can also specify a
date range as well as a department for the report.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 50

Once you display the report, you will see all accounts showing sum of all
transactions for that particular project. This report is similar to the trial balance
report but shows balances for a particular project only.

You can click on the account number to get the list of individual project
transactions for that account.

### 2.9 Chart of Accounts

A chart of accounts is required before you can start recording any accounting
transaction. When you create your company dataset in SQL-Ledger you have to
select one of the provided samples of chart of accounts. Later on you can modify
the initial chart of accounts according to your business needs.
The **’System–Chart of Accounts’** menu is used to manage the chart of


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 51

accounts. Here you can add new accounts, change existing ones or delete the
unwanted ones which have not been used in any transaction from the sample
chart of accounts.

The **’System–Chart of Accounts–List Accounts’** report shows the exist-
ing chart of account.

You can click on the account number to open the account in detail form
where you can make changes to the account. You can safely change the account
number at any time to reorganize your accounts. All transactions booked onto
the account will remain linked to it and will reflect the new number after change.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 52

#### 2.9.1 Heading accounts

All accounts in SQL-Ledger must be defined either as ’Heading’ or ’Account’. The
’Heading’ accounts help you divide your various accounts into groups. ’Head-
ing’ accounts are mainly for organizational purposes and are used to subtotal
groups of accounts in the income statement or balance sheet. You cannot record
transactions directly with heading accounts.

#### 2.9.2 Account types

The ’Account Type’ sets the accounting purpose for each account. Accounts
marked as ’Contra’ accounts are shown with reversed amounts in the trial balance.
Summary accounts are used to record transactions for accounts receivable,
accounts payable and inventory. If you mark an account as a summary account,
it will be included in the selection drop-down menus available when you process
accounts receivable and accounts payable transactions, or when you set up new
inventory item.

#### 2.9.3 Marking accounts

Here is how you mark accounts when adding or changing them:

1. When you mark an account to be included in the drop-down menus ’AR’,
    ’AP’ or ’Tracking Items’ (parts, assemblies, direct labor) and ’Non-tracking
    Items’ (services), it will be included in the respective modules.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 53

2. Marking ’Lineitem’ will make the account available as an income or expense
    account in AR and AP transactions.
3. Marking ’Payment’ will show that account for recording the receipt or
    payment of transactions.
4. Accounts marked to be included for ’Income’, ’COGS’ / ’Expense’ or ’Tax’
    under ’Tracking Items’ and ’Non-tracking Items’ will become available in
    the corresponding drop-down menus when you set up new goods and ser-
    vices under ’ **Goods & Services–Add...** ’.
5. In SQL-Ledger each tax account has a tax level which can be defined in
    ’ **System–Taxes** ’ for automatic calculation. Tax accounts can also be used
    for other purposes like commission fees.

#### 2.9.4 Mandatory default accounts

There are six default accounts in SQL-Ledger:

1. Income,
2. Expense
3. inventory
4. Foreign exchange gain
5. Foreign exchange loss
6. Cash over/short account.

You will also find them in the ’ **System–Defaults** ’ menu and once they have
been set up accordingly they cannot be deleted. You must also have at least
one account for accounts receivable and one for accounts payable, in order for
SQL-Ledger to be able to keep track of any outstanding amounts in the balance
sheet.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 54

#### 2.9.5 GIFI

GIFI stands for ’General Index for Financial Information’. GIFI codes can be
created and linked to accounts in chart of accounts. You can add or change GIFI
codes just like standard chart of accounts using ’ **System–Chart of Accounts–
Add GIFI** ’ and ’ **System–Chart of Accounts–List GIFI** ’ as shown with screen
shots below.

GIFI accounts can be used to re-group the regular accounts for reporting
purposes. All financial reports can be displayed with regular accounts or with gifi
accounts.

### 2.10 Templates

Print forms for invoices, orders, quotations and financial reports are defined as
templates. This makes it easy to customize these forms and reports according
to your requirements. Templates can be in html, LATEX or text format. These


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 55

templates can be accessed through the **’System’** menu.

#### 2.10.1 Editing Templates

Templates can be edited directly through SQL-Ledger user interface. When you
click on a template, it is displayed with an ’Edit’ button at the end of the screen.
Clicking the ’Edit’ button will open the template in a text box where it can be
edited and saved.

```
Here this template is opened for editing.
```

##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 56

#### 2.10.2 Template Variables

SQL-Ledger inserts actual data into the templates by using template variables.
Template variables are enclosed within <% and %>.
Here are some template variables to give you an idea. A simple way to view
all these template variables and understand their usage is to go through existing
sample templates which you can find in the ’ **System** ’ menu.

<%name%>
<%address1%>


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 57

<%address2%>
<%city%>
<%state%>
<%zipcode%>
<%country%>
<%contact%>
<%invnumber%>
<%invdate%>
<%duedate%>
<%ordnumber%>
<%employee%>
<%shippingpoint%>
<%shipvia%>
<%runningnumber%>
<%number%>
<%description%>
<%deliverydate%>
<%qty%>
<%unit%>
<%sellprice%>
<%discountrate%>
<%linetotal%>

#### 2.10.3 Template control commands

The template processing engine in SQL-Ledger allows simple conditional state-
ments and loops. Examples of these are described below:

**2.10.3.1 ’if’ is used to print a column data conditionally**

<%if contact%>
<br><%contact%>
<br>
<%end contact%>

<%if taxincluded%>
<th colspan=7 align=right>Total</th>
<td colspan=2 align=right><%invtotal%></td>
<%end taxincluded%>

<%if not taxincluded%>
<th colspan=7 align=right>Subtotal</th>
<td colspan=2 align=right><%subtotal%></td>


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 58

<%end taxincluded%>

<%if paid%>
<tr>
<th colspan=7 align=right>Paid</th>
<td colspan=2 align=right>- <%paid%></td>
</tr>
<%end paid%>

**2.10.3.2 ’for’ loop to print all lines on an invoice**

<%foreach number%>
<tr valign=top>
<td align=right><%runningnumber%>.</td>
<td><%number%></td>
<td><%description%></td>
<td><%deliverydate%></td>
<td align=right><%qty%></td>
<td><%unit%></td>
<td align=right><%sellprice%></td>
<td align=right><%discountrate%></td>
<td align=right><%linetotal%></td>
</tr>
<%end number%>

<%foreach tax%>
<tr>
<th colspan=7 align=right><%taxdescription%> on <%taxbase%> @ <%taxrate%> %</th>
<td colspan=2 align=right><%tax%></td>
</tr>
<%end tax%>


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 59

#### 2.10.4 Three type of templates:

**2.10.4.1 HTML Templates**

HTML templates are easier to modify because of the wide spread knowledge
of HTML. Only basic HTML knowledge is required to edit HTML templates.
The screen shot below shows the html template for sales invoice.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 60

The letterhead.html template is included in all other templates. You can for-
mat it to print your company name and other header information in a consistent
way across all the templates.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 61


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 62

**2.10.4.2 LATEX Templates**

LATEX templates are bit more complex to understand and modify, but are the
most powerful tool to generate printed documents in pdf or postscript format.
See 2.10.5 for a basic introduction to LATEX.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 63

As in html templates, the letterhead.tex template allows you to define your
company letter head for all templates in a consistent way.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 64

**2.10.4.3 Text Templates**

Text templates are used only for Point-of-Sale receipts printing. These tem-
plates allow you to print on 40 character receipt printers.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 65


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 66

#### 2.10.5 An Introduction to LATEX

LATEX is a complete collection of software tools to create high quality print doc-
uments like invoices, purchase orders etc. LATEX is included with most Linux
distributions. In the Red Hat distribution LATEX can be installed with the com-
mand ’yum install tetex’.

- In the Debian distribution it can be installed with ’apt-get install latex’.
- For FreeBSD, you can install the teTex port from /usr/ports/print/teTEX.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 67

LATEX might seem overwhelming to a newcomer but it is really a simple toolkit to
use for customizing the SQL-Ledger templates. In this very short introduction of
Latex, we shall go through the basic document format and its use in SQL-Ledger.
Here is ’Hello world!’ in LATEX.

**2.10.5.1 Create a text file (hello.tex) in your home folder with following
text:**

\documentclass[a4paper,11pt]{article}
\begin{document}
Hello world!
\end{document}

**2.10.5.2 Compile this tex file into dvi file and use xdvi to view it:**

latex hello.tex
xdvi hello.dvi

**2.10.5.3 You can also convert it to pdf:**

pdflatex hello.tex
xpdf hello.pdf

#### 2.10.6 Structure of a LATEX Document

Latex commands start with a backslash (\). Parameters can follow the command.
Optional parameters are enclosed in [ ] while mandatory ones are enclosed in {
}. { } can also be used to terminated a command mixed within some text (to
make it easier for the compiler to understand the command). Special characters
in latex (#, $, %, ^, &, _, {, }, ~) are escaped with \ except for the \ character
itself which is used to break a line. To use literal backslash (\) you can use the
special command $\backslash$.
Single line comments start with % while multi-line comments can be enclosed
between \begin{comment} and \end{comment} structure.
Every latex document starts with \documentclass with parameters ([a4paper,11pt]{article})
following it.

### 2.11 Goods & Services

All businesses sell some goods and services to generate revenue. You need to
define the goods and services related to your business before you can start creat-


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 68

ing invoices, orders or quotations. In SQL-Ledger goods and services have been
categorized into following entities:

1. Parts are something which you keep in the inventory and want to track
    their on-hand quantity.
2. Services are something which you provide to your customers or buy from
    vendors. Services are not ’stored’ somewhere and you do not track their
    on-hand quantity.
3. Assemblies are made up from parts, services and labor/overhead. This
    feature is used by manufacturing companies. When you build an assembly
    using the ’ **Goods & Services–Stock Assembly** ’ menu (see 2.11.4) all
    its associated parts are removed from inventory and the new assemblies
    are added to stock. When you sell an assembly COGS for parts and cost
    of services is recorded.
    Important note: Assemblies cannot be purchased and can only be sold.
4. Labor/overhead can be used to allocate the cost of labor or manufacturing
    overhead to the assemblies.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 69

#### 2.11.1 Parts

Parts are tangible items you keep in your stock. You purchase them from your
vendors and sell them to your customers for profit or you use them in an assembly.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 70

#### 2.11.2 Services

Services are intangible items which you sell and/or purchase.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 71

#### 2.11.3 Labor/Overhead

Labor/overhead items are used to allocate labor/overhead charges to an assembly
in a manufacturing business.

#### 2.11.4 Assemblies

An assembly is composed of components which are individual parts in the inven-
tory or other sub-assemblies. Assemblies in SQL-Ledger allow you to manage
your manufacturing process. Work flow for using assemblies is as follows:


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 72

1. Define assemblies. ’ **Goods & Services–Add Assembly** ’.
2. Build assemblies. **’Goods & Services–Stock Assembly** ’. Individual parts
    are removed and assemblies are added to the stock inventory.
3. Sell assembly items like any other item.

Please note that you cannot buy items defined as assemblies.

**2.11.4.1 Define assemblies**

An assembly is just like any other inventory item with the additional information
about its components. You define new assemblies using **’Goods & Service–Add
Assembly** ’.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 73

**2.11.4.2 Stock assemblies**

This option reduces the quantities of the components and increases the on-hand
quantity of the assemblies. COGS is not recorded at this point.
COGS for the assembly is recorded from individual components when you sell
the assembly. FIFO allocation also occurs at the time of sale.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 74

**2.11.4.3 Reports**

The **’Goods and Services–Reports–Stock Assembly’** menu gives you a list of
your ’Stock Assembly’ actions. This report lists the parts taken out of assembly
as well as assemblies built.

Following is a sample summary report of the ’Stock Assembly’ action. It
shows each action with its reference and date and other information.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 75

Following is a detail report of the ’Stock Assembly’ action. It shows each
assembly and its components which have been updated through the ’Stock As-
sembly’ action.

The **’Goods and Services–Reports–Assemblies’** menu gives you a list of
all or selected assemblies with their components. You can narrow down your
assemblies list by specifying search criteria.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 76

```
Once you click the ’Continue’ button above the following report is displayed.
```
The **’Goods and Services–Reports–Components’** menu gives you a list
ordered by part number and the assembly in which it is used.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 77

**Work Order**
You can print a work order for sales orders. A work order lists all component
parts required to fulfill a given order of assembly items.

#### 2.11.5 Groups

Groups are used to group together related parts, services and assemblies. You
can filter parts and services in the various ’ **Goods & Services–Reports** ’ by
selecting a group on the search screen.
Click on ’ **Goods & Services–Add Group** ’ menu to add a new group.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 78

To edit existing group, you will display list of existing groups using ’ **Goods
& Services–Reports–Groups** ’ and then click on the group name to edit that
group.

```
Click on the group name in the list above and it will be opened for change.
```
**2.11.5.1 Groups as POS buttons**

Groups also have another useful functionality. When you check the POS button
box while adding or changing a group, they will also appear as buttons on the
POS (point-of-sale) module screen making it easier to select items within each
group as shown below.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 79

**2.11.5.2 Subgroups**

You can also define subgroups. To define a subgroup you type of the name of
the group followed by a ’:’ and then the name of the subgroup. You can filter
certain reports with group or its subgroup.

#### 2.11.6 Pricegroups

SQL-Ledger has very flexible pricing mechanisms. For example:

1. You can define customer specific prices for each part.
2. You can define quantity breaks. For example, if someone buys 10 units
    instead of 1, he/she can automatically get a lower price.
3. And you can specify start and end dates to offer a special price during, for
    example, Christmas season.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 80

Price groups take this concept further and allow you to define ’groups’ of special
prices. Let us say you sell to distributors, dealers and end-users. Each of these
groups of customers gets tiered prices. There are three steps you need to take
to use price groups:

1. Create your price groups e.g. distributor, dealer and end-user using **’Goods**
    **& Services–Add Pricegroup** ’ menu.
2. Define item prices for these price groups. To do this, open the item for
    editing and select the price group. Then set the price according to the price
    group tier. Leave the customer column blank. Repeat this for all items.
    Clicking ’Update’ will allow you to set prices for multiple pricegroups for a
    single item.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 81

3. Open the customer record for editing and set the applicable price group
    for that customer.

### 2.12 Goods & Services Reports

Here we explain all reports under the **’Goods & Services–Reports’** menu briefly.

#### 2.12.1 All Items

This report can be used to view a list of all items which include parts, services,
labor/overhead and assemblies. You can optionally select to view invoices or
orders which have been created for each item.

#### 2.12.2 Parts

This report is similar to the all items report above but only shows parts or tangible
items for which you track on-hand quantity in your business.

#### 2.12.3 Requirements

This report will show you what you need to buy based upon the following factors:
On-hand quantity
Sales orders
Purchase orders


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 82

#### 2.12.4 Services

This report is similar to the all items report above, but only shows services.

#### 2.12.5 Labor/Overhead

This report is similar to the all items report but only shows labor/overhead items.

#### 2.12.6 Groups

This report will show you all the groups you have defined for your various goods
and services.

#### 2.12.7 Pricegroups

This report will show you all the price groups you have defined.

#### 2.12.8 Assemblies

This report will show you all the assemblies you have defined.

#### 2.12.9 Components

This report will show you all the components which have been used in your
assemblies.

#### 2.12.10Stock Assembly

This report will show you the log of stock assembly actions.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 83

### 2.13 Warehouses

Warehouses are optional and can be used to manage your inventory at more
than one physical place.
Note: Once you have defined warehouses, these are no longer optional and
you cannot post invoice or a transfer without specifying a warehouse.

#### 2.13.1 Adding warehouses

You can add, change or delete warehouses through the ’ **System–Warehouses** ’
menu.

#### 2.13.2 Default warehouse

You can specify a default warehouse for a user through ’ **HR–Employees** ’ menu.
This way user is restricted to his/her department for transaction entry or reports.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 84

#### 2.13.3 Using warehouses

The warehouse drop down menu is enabled on transactions screens once you
define at least one warehouse. When you purchase goods, quantity is added to
the specified warehouse. When you sell goods, quantity is subtracted from the
specified warehouse.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 85

#### 2.13.4 Warehouse transfers

You can move inventory between warehouses by using the ’ **Warehouses–Add
Transfer** ’ menu.

#### 2.13.5 Transfer Reports

The ’ **Warehouses–Reports–Transfers** ’ report shows a list of all transfers. On
the search screen you can select conditions to see only transactions of your
interest or just click ’Continue’ to display all transactions.
’Summary’ displays a list of of transactions and ’Detail’ display all items in
each transaction. You can click on the transfer number hyper link to edit the
transfer.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 86

#### 2.13.6 Warehouse Onhand Report

The **’Warehouses–Reports–Onhand’** report gives you inventory on-hand for
all warehouses or for a particular warehouse.

As you can see this report shows the onhand quantity of selected items at
each warehouse. This report can be sorted on item number so that you can
quickly see the on-hand quantity of a particular item at each warehouse.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 87

You can click on any item number to display the activity for that item as
shown below.

#### 2.13.7 Activity Report

**’Warehouses–Reports–Activity’** gives you a report of all activity of a partic-
ular item or all items. Select a warehouse to see the activity in that particular
warehouse. Activity report shows all the activity from purchase invoices, sales
invoices, shipped purchase orders, shipped sales orders and transfers.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 88

### 2.14 Languages

Language feature (accessible through ’ **System–Language** ’ menu) of SQL-Ledger
can be used for four main purposes:

1. You can define alternate descriptions, in a foreign language, for parts,
    services and groups (’ **Goods & Services–Translations** ’). This way you
    can send, for example, invoices to your customers with the description of
    your goods and services in their native language.
2. You can also translate the complete alternate set of templates which is cre-
    ated when you add a new language (See 2.10). This way you can send, for
    example, invoices to your customers where the standard template content
    is translated into a foreign language. This can be used in combination with
    the alternate descriptions for parts, services and groups mentioned above,
    or on its own to define a particular set of documents for a particular cus-
    tomer or market segment.
3. You can translate your chart of accounts if you want to be able to print
    your ’Balance Sheet’ and/or ’Income Statement’ in a foreign language using
    ’ **System–Chart of Accounts–Translations** ’ menu.
4. You can also translate the balance sheet and income statement templates
    for a foreign language using ’ **System–html templates–Income State-**
    **ment** ’ and **’System–html templates–Income Statement’**

To define a new language, use the **’System–Language’** menu. An existing
list of languages (if any) will be displayed with the ’Add Language’ button at
the bottom of the report. When you add a new language, SQL-Ledger adds a
complete alternate set of templates for that language.

Once you have defined a new language, you can see it on a drop-down menu
in the invoice, order and quotation print options area.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 89

### 2.15 Translations

Once you have defined a language, you can add translations for certain things
like:

1. Chart of accounts
2. All items
3. Groups
4. Projects

To add a translation, use the **’Translations’** sub-menu under the respective
menu.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 90

### 2.16 Taxes

Defining and using taxes is a four step process:

#### 2.16.1 Define the tax accounts in chart of accounts

You can create or edit tax accounts in the chart of accounts using the ’ **System–
Chart of Accounts** ’ menu and by marking the ’Tax’ checkbox under the relevant
group as shown below.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 91

#### 2.16.2 Define tax percentages

You set percentages for each tax using the ’ **System–Taxes** ’ menu. If the tax
rate changes you enter the last date into the ’Valid To’, click ’Update’ and enter
the new rate in the new line.

#### 2.16.3 Mark Items/Services as taxable

You mark each part or service as taxable during the ’add’ or ’edit’ process. You
do this using the ’ **Goods & Services** ’ menu. Once a part or a service has been
sold, the tax account should not be changed.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 92

#### 2.16.4 Mark Customers/Vendors for applicable taxes

Tax will not be calculated for your customers or vendors unless you mark them
as taxable. You do this using **’Customers’** or **’Vendors’** menu.

### 2.17 Data import from other applications

Sometimes you need to import your sales data which was produced elsewhere
into SQL-Ledger. You might have a web store where you download your daily
sales in CSV format and want to import it into SQL-Ledger. Or you are just
moving to SQL-Ledger from your legacy accounting software and want to move
all existing data from your old software to SQL-Ledger.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 93

In SQL-Ledger, we can import data for almost everything as shown in the
image above. The following sections provide detailed information about the steps
to take for importing data from CSV text files into SQL-Ledger.

#### 2.17.1 Sale invoices

Sales invoices can be imported from CSV text files.

**2.17.1.1 Format your data**

Here is a sample of sales invoice import data. You prepare data in this format
and save it in a text file. The last column AR is accounts receivable account
number which is 1100 in UK chart of accounts.
If your data contains invoices with more than one item, repeat the row with
same invoice header information and change the item number and price infor-
mation. SQL-Ledger will import all these rows as a single invoice. (See invoice
number A100 above)
For list of additional data columns that can be imported see step 4.
invnumber,transdate,duedate,customernumber,curr,invoicedescription,partnumber,
qty,sellprice,employeenumber,AR,department,warehouse


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 94

A100,10/12/2008,10/30/2008,AE001,GBP,Invoice description comes here,B001,10,102,E-001,1100,
HARDWARE,LONDON
A100,10/12/2008,10/30/2008,AE001,GBP,Invoice description comes here,F003,6,69,E-001,1100,
HARDWARE,LONDON
A101,10/12/2008,10/31/2008,CP002,GBP,Test description,F003,2,32,E-002,1100,SERVICES,PARIS
A102,10/13/2008,11/1/2008,ER003,GBP,Sale of goods,T007,6,12,E-003,1100,SERVICES,LONDON
A103,10/14/2008,11/2/2008,SP007,GBP,Sale,K001,12,32,E-004,1100,HARDWARE,PARIS

**2.17.1.2 Upload and preview**

Use the ’ **Import–Sales Invoices** ’ menu option to upload your file into SQL-
Ledger. You will be shown what will be imported before actual import is done.
At this point you can check and uncheck the invoices to be imported.

**2.17.1.3 Confirm data import**

When you click the Import Sales Invoices button, invoices will be imported. You
will be show which invoices were imported successfully.

**2.17.1.4 Additional data which can be imported**

The sample CSV file provided above contains only the most commonly used
columns. Here is the complete list.
transdate
invnumber
customernumber
curr
duedate
employeenumber
ordnumber
quonumber
datepaid


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 95

```
shippingpoint
shipvia
waybill
terms
notes
intnotes
language_code
ponumber
cashdiscount
discountterms
partnumber
description
sellprice
discount
qty
unit
serialnumber
projectnumber
deliverydate
AR
taxincluded
```
#### 2.17.2 Receipts and Payments

You can import payments and match them to invoices using the ’ **Import–
Payments** ’ menu. The following points should be kept in mind.

1. Payments are matched first on the Invoice DCN column and then, if no
    match is found, on the payment amount.
2. Both AR and AP invoices are matched with payments.
3. The amount matched is calculated as debit minus credit.

**2.17.2.1 Format your data**

Create or format the data in a CSV file with structure similar to the one given
below.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 96

datepaid,memo,debit,credit,dcn
2008/11/03,"payment ref 2121",,38.76,
2008/10/04,"cash payment",,527.5, 2008/10/10,"CC Receipt",,243.08,
2009/11/01,"Payment matched by DCN",,1401.72,1122

**2.17.2.2 Upload and perview**

The import script will read the CSV file and match the payments to AR or AP
invoices first on the DCN Number and then on the invoice due amount, if needed.
In this example, one AP invoice is matched on the amount and the other
one is matched on the DCN number. The other two are AR invoices which are
matched on the amount.

**2.17.2.3 Confirm data import**

Once you click ’Import Payments’, payments are imported and applied to the
matched invoices.

**2.17.2.4 Advanced receipts/payments import**

1. You can easily change the script to match the payments on other invoice
    columns like invoice number. The procedures to modify are located in ’sub
    payments’ in ’SL/IM.pm’ and ’sub im_payment’ in the ’bin/mozilla/im.pl’
    file.
2. To match payments only to AR (or AP) invoices, change the UNION
    queries in the ’SL/IM.pm’ file to select invoices from AR or AP only as
    required.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 97

#### 2.17.3 AR/AP Transactions

You can import both AR and AP transactions.
For AR transactions, format your data using the following sample:

invnumber,customernumber,transdate,amount,description,notes,source,memo
00003,AE001,10-11-07,2030,"desc1","notes1","source1","memo1"
00004,CP002,07-12-07,3213,"desc1","notes2","source2","memo2"
00005,SP007,09-12-07,-200,"desc1","notes3","source3","memo3"

For AP transactions, format your data using the following sample:
invnumber,vendornumber,transdate,amount,description,notes,source,memo
00003,CB001,10-10-08,2030,"desc1","notes1","source1","memo1"
00004,ES002,10-12-08,3213,"desc2","notes2","source2","memo2"
00005,SA003,12-12-08,-200,"desc3","notes3","source3","memo3"

#### 2.17.4 General Ledger

This feature will help you to move your data from most of the accounting software
to SQL-Ledger in just a few easy steps:

**2.17.4.1 Format your data**

Format your data according to the following sample. Keep in mind that:

1. The import script will create one GL transaction for each unique ’reference’
    number.
2. There can be any number of lines (rows) in each transaction.
3. The imported account must also exist in the SQL-Ledger chart of accounts.
4. Debits and credits must be equal before the CSV file can be imported.

reference,transdate,description,notes,accno,debit,credit,source,memo
GL001,01-20-2008,"Paid for training,support",Next session in 2009,8203,124,0,23211,new
hiring
GL001,01-20-2008,"Paid for training,support",Next session in 2009,1230,0,124,23211,new
hiring
GL002,10-19-2008,"Overdue pymt for inv 11,12,13",,1230,204,0,"11,12,13",
GL002,10-19-2008,"Overdue pymt for inv 11,12,13",,1102,0,204,"11,12,13",
GL003,11-20-2008,Invalid transaction for testing,This account is not in chart,00121,0,255,
source2,memo2


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 98

**2.17.4.2 Upload and preview**

Use the **’Imports–GL Transaction’** to load the CSV file into SQL-Ledger. The
import script will show ’****’ in the ’Account Description’ column, if the row
to be imported doesn’t contain a valid account number. Only account numbers
that exists in the SQL-Ledger chart of accounts are valid account numbers.

**2.17.4.3 Confirm data import**

Click Import GL to finish the import script. Transactions successfully imported
will be show on the next page.

#### 2.17.5 Customers and Vendors

Customer and Vendor import is similar (except for the number column which is
either ’customernumber’ or ’vendornumber’).
Prepare your data file using the sample text provided below. (Change cus-
tomernumber to vendornumber for vendor import)

customernumber,name,firstname,lastname,contacttitle,phone,fax,email,notes,address1,address2
,city,state,zipcode,country
001,Ledger123,Armaghan,Saqib,Consultant,,,saqib@ledger123.com,"These are, just, sample
notes",,,London,,"AA7 8BB",UK

#### 2.17.6 Parts

**2.17.6.1 Format your data**

Format your data according to following sample format. Please note that:

1. The import procedure assigns a unique parts_id to each part imported or
    group created.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 99

2. Duplicates are not allowed and duplicate check is done on partnumber.

partnumber,description,unit,partsgroup,listprice,sellprice,lastcost,rop,bin,image,drawing,
notes
B002,"Brush Set",NOS,brush,9.99,9.99,7,150,TOP,noimage,brush.jpg,notes about brush set
D010,"Deluxe Hand Saw",NOS,SAW,17.99,17.99,16,50,TOP,saw.jpg,nodrawing,notes about hand saw
D011,"Digger Hand Trencher",NOS,Picks & Hatchets,18.99,18.99,15,200,TOP,,nodrawing,notes
about hand saw

**2.17.6.2 Upload and preview**

To start the import process, click ’ **Data Import–Parts** ’ in the menu. The
following page will be displayed. Click ’Browse...’ to select your CSV file, mark
the taxes applicable and select the account links (The defaults are usually enough)
Click ’Continue’ when done. You will be presented with the following screen. On
this screen you can mark the parts to be imported by checking or un-checking
the check-box on each line.
Please note:

1. Any parts which are already in SQL-Ledger (based on ’partnumber’) will
    not imported. (You will not see a check-box with them)
2. Parts ’groups’ which are new will be added. These are marked by a ’+’
    sign after group name.


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 100

**2.17.6.3 Confirm data import**

Click ’Import Parts’. Your CSV file will be processed and parts will be imported.
Any new groups will also be added. You will see an output like the following:

#### 2.17.7 Vendor price list

**2.17.7.1 Format your data**

partnumber,vendornumber,vendorpartnumber,lastcost,curr,leadtime
B001,CB001,V-CB001,10,GBP,15 B002,ES002,,14,GBP,45 M004,SA003,,21,GBP,30

**2.17.7.2 Upload and preview**

To start the import process, click ’ **Import–Vendor Price List** ’ in the menu,
specify your CSV file with the ’Browse’ button and click the ’Import Parts Ven-
dors’ button. The following page will be displayed. Here you can un-check the
rows which you do not want to import. Rows with an invalid ’vendor number’
or ’partnumber’ will not have the check-box.

#### 2.17.8 Customer price list

**2.17.8.1 Format your data**

partnumber,customernumber,pricegroup,pricebreak,sellprice,validfrom,validto,curr
B001,AE001,PG1,10,11,03-01-2008,,GBP
B002,BP011,,20,12,,03-01-2009,GBP
M004,CP002,,15,20,03-01-2008,03-05-2008,GBP
D08,CP002,test,25,25,,,GBP

**2.17.8.2 Upload and preview**

To start the import process, click ’ **Import–Customer Price List** ’ in the menu,
specify your CSV file with the ’Browse...’ button and click the ’Import Parts
Customers’ button. The following page will be displayed. Here you can un-check


##### CHAPTER 2. SETTING UP YOUR BUSINESS ON SQL-LEDGER 101

the rows which you do not want to import. Rows with an invalid ’customer
number’ or ’partnumber’ will not have the check-box.

#### 2.17.9 Chart of accounts

**2.17.9.1 Format your data**

1. Prepare your chart of accounts in your spreadsheet software according to
    the sample given below.
2. Upload the chart CSV file using ’ **Import–Chart** ’ menu option.
3. Check/un-check the accounts to be imported and click ’Continue’ to import
    the selected accounts.

accno,description,charttype,category,link
1000,"CURRENT ASSETS",H,A,
1060,"Checking Account",A,A,AR_paid:AP_paid
1065,"Petty Cash",A,A,AR_paid:AP_paid
1200,"Accounts Receivables",A,A,AR
1205,"Allowance for doubtful accounts",A,A,
1500,"INVENTORY ASSETS",H,A,
1520,"Inventory / General",A,A,IC
1530,"Inventory / Aftermarket Parts",A,A,IC
1800,"CAPITAL ASSETS",H,A,


# Chapter 3

# Running your business on

# SQL-Ledger

In this chapter you will learn how to use different SQL-Ledger modules to process
your business transactions. Each module has been explained in detail with screen
shots and explanation.

### 3.1 AR

AR stands for ’Accounts Receivable’. The AR module is used to record your sales
to customers. You can record your sales in two ways:

1. ’ **AR–Add Transaction** ’ is a simplified way to book your sales and receipts
    using pre-defined accounts from chart of accounts. This method is quick
    and requires no setup of goods or services. Only in some circumstances it
    might require the adjustment of the chart of accounts to suit your individual
    needs.
2. ’ **AR–Sales Invoice** ’ is the standard way to record sales and receipts. In
    a sales invoice you will specify the goods and/or services you have sold to
    your customer. This method requires the setup of goods and services using
    the ’ **Goods & Services** ’ menu (see 2.11). You can print a sales invoice
    and email it to your customer. If you are managing inventory, you need to
    use this method to reduce the inventory when you sell something.

Both methods can be mixed and matched based upon the nature of your trans-
action.

##### 102


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 103

#### 3.1.1 AR Transaction

The **’AR–Add Transaction’** menu is used to create a simple AR transaction.
These transactions allow you to record your sales on general ledger accounts
without creating an invoice.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 104

#### 3.1.2 Sales Invoice

Sales invoices are created using the **’AR–Sales Invoice’** menu. The only manda-
tory columns in the header section of this screen are ’Customer’ and ’Invoice
Date’. All other columns in the header section can be left blank.
Your invoice can contain multiple items in the detailed section of the invoice
(parts, assemblies, services and labor etc.). When you enter the article number
or description of one of your items and click ’Update’, the master data for that
item (article number, description, price and unit) is displayed in the current row.
You can then enter the quantity you want to sell and click ’Update’ or press
return. A new line appears and you can add another item and so on. If the
article number or description you enter is not found in the database, SQL-Ledger
asks if you want to add a new part or service to your master data. This way


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 105

you can enter any number of items, both existing and new, in your sales invoice.
If you are uncertain of the article number or description, you can enter ’%’ and
click ’Update’ or press return. SQL-Ledger will then list all available items and
you can select the one you want by marking the appropriate checkbox.

By default only ’item number’, ’description’, ’qty’, ’unit’, ’price’ and ’dis-
count’ are shown on each line item. You can display additional fields for extended
information input on each line item. To do this, just mark the check-box next
to the ’Description’ column of the line item or heading and click ’Update’. Now
the invoice form is displayed with extended line items as shown below.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 106

#### 3.1.3 Credit invoice and credit note

Credit invoices are used to record a sale return which was recorded earlier with
a sales invoice. A credit invoice will add the items you sold earlier back to
the inventory for re-sale as well as update your accounts receivable and sales
accounts.
Credit notes are used to record a sale return without creating a credit invoice.
A credit note is typically used to record the reversal of an ’AR Transaction’ (see
3.1.1 above), though it can also be used to reverse all or part of a sales invoice,
but be aware that inventory is not added back to your stock with a credit note.
So credit note is a good tool to reverse any service sale, but not for reversing
tangible a goods’ sale.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 107

**3.1.3.1 Creating a credit invoice**

Use the ’ **AR–Credit Invoice** ’ menu to create a credit invoice. Creating a credit
invoice is similar to creating a sales invoice. See 3.1.2 for details on how to do
that.

**3.1.3.2 Creating a credit note**

Use the ’ **AR–Credit Note** ’ menu to create a credit note. Creating a credit note
is very similar to creating an AR transaction. See 3.1.1 for details on how to do
that.

**3.1.3.3 Adjusting a credit note or a credit invoice**

Once you have an open invoice as well as a credit note or credit invoice for a
certain customer, you can adjust these against each other. To do this:

1. Use the ’ **Cash–Receipt** ’ menu to select the customer and click the ’Up-
    date’ button. This will list all open invoices as shown below.
2. Mark the invoices/ transactions you want to adjust and click ’Update’.
    When the amounts of both open invoice and credit invoice are equal and
    thus the total amount is zero, the ’Amount’ field in the header section will
    remain empty. For your reference you can put something like ’adjustment’


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 108

```
in source column.
```
3. Now you just click ’Post’ and the credit invoice will be adjusted against
    the open sales invoice.

### 3.2 AR reports

#### 3.2.1 Transactions report

The AR transactions report lists all open AR transactions and invoices. You can
specify different criteria and select / de-select any columns you want to display in


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 109

the report. If you mark the ’Closed’ checkbox, the report will also list all closed
AR transactions and invoices.

When you click the ’Continue’ button after specifying any chosen criteria,
your report is displayed. The ’Summary’ report lists each invoice or transaction
on a single line as shown below.

```
The ’Detail’ report will also list the single debits and credits of each trans-
```

##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 110

action along with the related account number. In the detail report, a single
invoice or transaction is displayed on multiple rows. You can mark the ’Subtotal’
checkbox to subtotal and group this report by invoice number as shown below.

#### 3.2.2 Aging report

The aging report lists the outstanding balances of your customers and divides
them into predefined periods of time.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 111

The ’Summary’ aging report lists each customer with an outstanding balance
on a single row as shown below.

The ’Detail’ aging report also lists the single outstanding invoices for each
customer with their respective subtotal.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 112

#### 3.2.3 Reminders

Reminders can be printed or emailed to directly the customers. You can define
up to 3 levels of reminders. Level 1 being polite and level 3 being a bit harsh.
When you print or email a reminder, the respective reminder level is stored
in the database. The next time you print a reminder for the same customer, the
following level of reminder is already preset. You also have the option to change
the reminder level manually by clicking the ’Save level’ button.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 113

#### 3.2.4 Customer history reports

You can use the customer history reports to see exactly what your various cus-
tomers are buying. You can also filter the report for invoices, orders and quota-
tions between any date range, or on other selected criteria.

The ’Summary’ report for customer history will list the business activity
grouped by item as shown below.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 114

The ’Detail’ report for customer history will list the business activity by invoice
and individual item as shown below.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 115

### 3.3 Point of sales (POS)

The point of sales (POS) module allows quick invoicing at busy places like a shop
or a restaurant. The items and customers you have defined for your sales invoices
can also be used for POS invoicing. The only difference between POS invoice
creation and standard AR invoice creation is a simplified data entry screen and
a POS optimized receipts section.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 116

#### 3.3.1 Creating a POS invoice

Use the ’ **POS–Sale** ’ menu to create a new POS invoice. The screen shown
below is displayed. Here you select the customer and then add the items (parts
or services) which you sell to that customer.

Hint: If you sell to mainly walk-in customers and don’t want to create a customer
record for each walk-in customer then you can just add a customer with ’Walk-in’
as customer name.
Item groups are shown as buttons on the POS screen to make it easier to
select the item you want to sell. You first need to check the checkbox ’POS
Button’ of each individual group to see it as a button on the POS screen (see
2.11.5.1). You can then click this group button on the POS screen to display
all the items contained in that group and select the individual items you want to
sell.

#### 3.3.2 Viewing open invoices

In places like retail shops a POS invoice is created and closed in one step. In
places like a restaurant, there can be a considerable time period when an invoice


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 117

is created and when it is closed. In the later scenario, you create a POS invoice
when the customer has ordered his food. Once the customer has consumed the
food and is ready to pay, you simply locate the invoice and add the payment to
it.

To open a particular POS invoice, you view the open invoices using the
’ **POS–Sale** ’ menu and then click on the invoice number of your interest. In the
payment section, you can enter the payment received from the customer as well
as the account to credit (cash, credit card or something else).
If you enter a higher payment amount than the total amount of the invoice,
the rest of the payment amount will be shown as ’change’ which needs to be
returned back to the customer and the invoice will be closed with the payment
amount equal to the invoice amount.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 118

#### 3.3.3 Receipts

The receipts report shows all the receipts done so far with your POS module.
Use the ’ **POS–Receipts** ’ menu to view this report.

### 3.4 AP

AP stands for ’Accounts Payable’. The AP module is used to record purchases
from your vendors. You can record your sales in two possible ways:

1. ’ **AP–AP Transaction** ’ is a simplified way to book your purchases, ex-
    penses and payments using pre-defined accounts from the chart of ac-
    counts. This method is quick and requires no setup of goods or services.
    Only in some circumstances it might require the adjustment of the chart
    of accounts to suit your individual needs.
2. ’ **AP–Vendor Invoice** ’ is the standard way to record purchases. In a vendor
    invoice you can specify the goods and/ or services you have purchased from
    your vendor. This method requires the setup of goods and services using
    the ’ **Goods & Services** ’ menu. If you are managing your inventory, you
    need to use this method to increase the inventory when you buy something.

Both methods can be mixed and matched based upon the nature of your trans-
actions and business.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 119

#### 3.4.1 AP transactions

The **’AP–Add Transaction’** menu is used to create a simple AP transaction.
These transactions allow you to record your purchases and expenses on general
ledger accounts without creating a vendor invoice.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 120

#### 3.4.2 Vendor invoice

Vendor invoices are created using the **’AP–Vendor Invoice’** menu. The only
mandatory columns in the header section of this screen are ’Vendor’ and ’Invoice
Date’. All other columns in the header section can be left blank.
Your invoice can contain multiple items in the detailed section of the invoice
(parts, assemblies, services and labor etc.). When you enter the article number or
description of one of your items and click ’Update’, the master data for that item
(article number, description, price and unit) would be displayed in the current
row. You can then enter the quantity you want to sell and click ’Update’ or press
return. A new line will appear and you can add another item and so on.
If the article number or description you enter is not found in the database,
SQL-Ledger would ask if you want to add it as a new part or service to your
existing goods and services. This way you can enter any number of items, both
existing and new, in your sales invoice.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 121

If you are uncertain of the article number or description, you can enter ’%’
and click ’Update’ or press return. SQL-Ledger will then list all available items
and you can select the one you want by marking the appropriate checkbox.

By default only ’item number’, ’description’, ’qty’, ’unit’, ’price’ and ’dis-
count’ are shown on each line item. You can display additional fields for extended
information input on each line item. To do this, just mark the check-box next
to the ’Description’ column of the line item or heading and click ’Update’. Now
the invoice form will be displayed with extended line items as shown below.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 122

#### 3.4.3 Debit invoice and debit note

Debit invoices are used to record a purchase return which was recorded earlier
in a vendor invoice. A debit invoice will remove the items you purchased earlier
from your stock inventory as well as update your accounts payable and purchase
accounts.
Debit notes are used to record a sale return without creating a debit invoice.
A debit note is typically used to record reversal of an ’AP Transaction’, though
it can also be used to reverse all or part of a vendor invoice, but be aware that


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 123

inventory is not removed from your stock with a debit note. So debit note is
good tool to reverse any service purchases, but not for reversing tangible goods
purchase.

**3.4.3.1 Creating a debit invoice**

Use the ’ **AP–Debit Invoice** ’ menu to create your debit invoice. Creating a debit
invoice is similar to creating a vendor invoice. See 3.4.2 for details on how to do
that.

**3.4.3.2 Creating a debit note**

Use the ’ **AP–Debit Note** ’ menu to create a debit note. Creating a debit note
is very similar to creating an AP transaction. See 3.4.1 for details on how to do
that.

**3.4.3.3 Adjusting debit note or debit invoice**

Once you have an open vendor invoice as well as a debit note or debit invoice,
you can adjust them to each other. To do this:

1. Use the ’ **Cash–Payment** ’ menu to select the vendor and click the ’Update’
    button. This will list all open invoices as shown below.
2. Mark the invoices/ transactions you want to adjust and click ’Update’.
    When the amounts of both open invoice and credit invoice are equal and


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 124

```
thus the total amount is zero, the ’Amount’ field in the header section will
remain empty. For your reference you can put something like ’adjustment’
in the source column.
```
3. Now you just click ’Post’ and the debit invoice will be adjusted against the
    open vendor invoice.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 125

### 3.5 AP reports

#### 3.5.1 Transactions report

The AP transactions report lists all open AP transactions and invoices. You
can specify your search criteria and select / de-select any columns you want to
display in the report. If you mark the ’Closed’ checkbox, the report will also list
all closed AP transactions and invoices.

When you click the ’Continue’ button after specifying any chosen criteria,
your report is displayed. The ’Summary’ report lists each invoice or transaction
on a single line as shown below:


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 126

The ’Detail’ report will also list the single debits and credits of each trans-
action along with the related account number. In the detail report, a single
invoice or transaction is displayed on multiple rows. You can mark the ’Subtotal’
checkbox to subtotal and group this report by invoice number as shown below.

#### 3.5.2 Aging report

The aging report lists the outstanding balances of your vendors and divides them
into predefined periods of time.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 127

The ’Summary’ aging report lists each vendor with an outstanding balance
on a single row as shown below.

The ’Detail’ aging report also lists the single outstanding invoices for each
vendor with their respective subtotal.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 128

#### 3.5.3 Vendor history reports

You can use the vendor history reports to see exactly what you have purchased
from your different vendors. You can also filter the report for invoices, orders
and quotations between any date range, or on other selected criteria.

The ’Summary’ report for vendor history lists the business activity grouped
by item as shown below.

```
The ’Detail’ report for vendor history lists the business activity by invoice and
```

##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 129

individual item as shown below.

### 3.6 Cash

#### 3.6.1 Receipts

The **’Cash–Receipt’** menu is used to record receipts from your customers and
to adjust the outstanding balance. The second menu entry **’Cash–Receipts’**
allows you to enter receipts from multiple customers. Both have the same effect,
but the second one make data entry quicker when adding receipts from multiple
customers.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 130

**3.6.1.1 Receipt from a single customer**

There are two ways to record a receipt from a single customer:

1. If the invoice is paid at the time of sale, you can enter the receipt infor-
    mation in the footer section of the invoice screen whilst creating it. This
    way the invoice will immediately be considered closed when you post it.
2. If the invoice is paid later, you can use the **’Cash–Receipt’** menu to record
    the receipt for a particular customer. Using this method is advisable, as
    you do not need to edit the invoice to record the receipt. This method
    also allows you to record a one time receipt for multiple invoices.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 131

**3.6.1.2 Receipts from multiple customers**

The **’Cash–Receipts’** menu allows you to quickly record receipts from multiple
customers.

**3.6.1.3 Receipts report**

This report, accessible via ’ **Cash–Reports–Receipts** ’ menu, shows you all re-
ceipts on a selected bank account.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 132

#### 3.6.2 Payments

The **’Cash–Payment’** menu is used to record payments to your vendors and
to adjust the outstanding balance. The second menu entry **’Cash–Payments’**
allows you to enter payments to multiple vendors. Both have the same effect,
but the second one make data entry quicker when adding payments to multiple
vendors.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 133

**3.6.2.1 Payment to a single vendor**

There are two ways to record a payment to a single vendor:

1. If the invoice is paid at the time of purchase, you can enter the payment
    information in the footer section of the invoice screen whilst creating it.
    This way the invoice will immediately be considered closed when you post
    it.
2. If the invoice is paid later, you can use the **’Cash–Payment’** menu to
    record the payment to the invoices of a particular vendor. Using this
    method is advisable, as you do not need to edit the invoice to record the
    payment. This method also allows you to record a one time payment for
    multiple invoices.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 134

**3.6.2.2 Payments to multiple vendors**

The **’Cash–Payments’** menu allows you to quickly record payments to multiple
vendors.

**3.6.2.3 Payments report**

This report, accessible via ’ **Cash–Reports–Payments** ’ menu, shows you all
payments on a selected bank account.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 135

### 3.7 General ledger

The ’ **General Ledger** ’ menu is used to add manual debit and credit accounting
entries to selected accounts from your chart of accounts. You cannot post a
transactions until the total of debits is equal to the total of credits.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 136

#### 3.7.1 Add transaction

Use the **’General Ledger–Add Transaction’** menu to add a new general ledger
transaction. On this screen you can put some reference number in the ’Reference’
column. If you leave it blank, SQL-Ledger will assign the next number from the
scheme defined in the ’ **System–Defaults** ’ menu.

#### 3.7.2 Reports

The **’General Ledger–Reports’** menu is used to view all accounting journal
entries with the debits and credits on the related accounts. Initially this report
can be confusing because it shows not only the journal entries added using the
’ **General Ledger–Add Transaction** ’ menu described above, but also lists all
other accounting transactions posted from the AR, AP and Cash modules.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 137

```
The general ledger report can be sorted on any column.
```

##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 138


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 139

### 3.8 Recurring transactions

Recurring transactions allow you to auto-generate predefined invoices, transac-
tions and orders. This feature can be used for the following:

1. Recurring billing to a customer (for rent, web hosting, school fee, install-
    ment etc.)
2. Recurring billing from your vendor
3. Monthly orders to your vendors or from your customers.
4. Monthly payroll posting using general ledger recurring transactions.
5. Month-end adjustments and allocations.

#### 3.8.1 Scheduling

To schedule a recurring transaction, you need to start by first manually booking
the transaction you want to repeat. Once you have created the first model
transaction, you can edit it and click on the ’Schedule’ button at the bottom of
the screen. You will then be able to set the criteria for your recurring transactions.
SQL-Ledger will use your manually created transaction as model for the recurring
transactions and process it according to the individually chosen settings. To


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 140

automatically generate the next number for a given transaction, just leave the
’Reference’ field blank.

#### 3.8.2 Generating

When recurring transactions are due, you are reminded next time you login to
SQL-Ledger. With a single click you can then generate all recurring transactions
and print or email invoices and orders.

### 3.9 Currencies and exchange rates

You can define and use multiple currencies in SQL-Ledger.

#### 3.9.1 Defining currencies

To define a new currency use the ’ **System–Currencies** ’ menu. The currency
listed at the top will be your default currency. You can move the currencies up
and down using the arrows in the currency list.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 141

#### 3.9.2 Buying and selling in foreign currencies

When you want to create an invoice in a foreign currency, just change the currency
code in the currency drop-down box and enter the appropriate exchange rate.
If the exchange rate for a certain day has already been entered in a previous
transaction, SQL-Ledger will automatically show you the set exchange rate for
this currency. You can either choose to accept this rate or change it to something
else.

#### 3.9.3 Reports

You can view all reports in your base currency as well as in foreign currency. To
see the foreign currency used in a certain transaction, simply mark the ’Currency’
check-box.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 142

#### 3.9.4 Exchange rate difference

Any exchange rate difference that occurs between the time of sale or purchase
and the time of receipt or payment, will be automatically booked onto the for-
eign exchange gain or loss accounts that have been predefined in the ’ **System–
Defaults** ’ menu.

#### 3.9.5 Fund transfers in foreign currencies

If you want to transfer funds to or from a foreign currency account, you should
use the ’ **Cash–FX Adjustment** ’ module. Let’s assume that you want to transfer
100 GBP to your USD account and that the exchange rate is 1 GBP = 2.0289
USD (or reverse 1 USD = 0.4929 GBP). Then you should proceed as follows:

Or reverse, to transfer 100 USD to your GBP account, you should proceed
as follows:

### 3.10 Quotations and RFQs

You can use SQL-Ledger to send quotations to your customers or request quo-
tations from your vendors (RFQs). Later on you can convert the quotations to
sales orders and the RFQs to purchase orders.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 143

#### 3.10.1 Quotations

Use the **’Quotations–Quotation’** menu to add a new quotation for your cus-
tomer.

To get a report of existing quotations or to edit a quotation, use the **’Quotations–
Reports–Quotations’** menu. The search screen will be displayed. Here you can
specify conditions to filter the report or just leave it blank and click the ’Continue’


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 144

button to list all open quotations.

Once a quotation has been made, you can create a purchase order from
it. Creating a PO from a quotation will mark it closed. You can also close a
quotation by clicking on the ’Closed’ radio button and then saving it by clicking
on the ’Save’ button at the bottom of the screen.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 145

#### 3.10.2 RFQ

Use the **’Quotations–RFQ’** menu to add a new request for quotation from your
vendor.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 146

To get a report of the existing RFQs or to edit an RFQ, use the **’Quotations–
Reports–RFQs’** menu. The search screen will be displayed. Here you can specify
conditions to filter the report or just leave it blank and click the ’Continue’ button
to list all open RFQs.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 147

Once an RFQ has been made, you can create a purchase order from it.
Creating an PO from an RFQ will mark it closed. You can also close an RFQ
by clicking the ’Closed’ radio button and then saving it by clicking on the ’Save’
button at the bottom of the screen.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 148

### 3.11 Orders

SQL-Ledger has a very powerful orders management module which supports full
or partial shipping / receiving of orders along with complete inventory manage-
ment at multiple warehouses. The orders module can be used to create purchase
orders for your vendors and sales orders for your customers.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 149

```
Here are few points to remember:
```
1. When you create an invoice from an order, you cannot edit the quantities
    on the invoice screen or add or remove items. This is intended program
    function to keep invoices and orders correctly cross-referenced.
2. When you create an invoice from a partially received order, this order is
    marked closed and a new order with the same number, but only with the
    remaining quantities and with a new order date is created.

#### 3.11.1 Sales orders

Creating a sales order is often the first step you take when you sell goods and
services to your customers. You can:

1. Make a sales order.
2. Receive a sales order fully or partially using the ’ **Shipping–Ship** ’ menu.
3. Create a customer invoice from a partially or fully shipped sales order.
4. If warehouses are enabled, you can also ship goods from a particular ware-
    house.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 150

**3.11.1.1 Add a new sales order**

The **’Order Entry–Sales Order’** menu will display the following ’Add Sales
Order’ screen.

If you want to enter more information for each item you can mark the check-
box next to the ’Description’ column and then click ’Update’. Now each detail
line will span 5 lines where you can enter lots of other information for each item
you sell.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 151

**3.11.1.2 Editing a sales order**

To edit an existing sales order, you display a list of existing orders using the
**’Order Entry–Reports–Sales Order’** menu and click on the sales order number
to edit that particular sales order.

**3.11.1.3 Creating a quotation or customer invoice from a sales order**

Once you have saved a sales order, you can open it again in editing mode and
use it to create a new quotation. When you have shipped quantities you can also
directly create a sales invoice from the sales order. (Also see below to see how
to use the ’ **Shipping** ’ menu to partially ship a sales order.)

**3.11.1.4 Shipping a sales order**

There are two ways to ship a sales order.

1. Open the sales order and click the ’Customer Invoice’ button. The sales
    order will then automatically be shipped in full, marked ’closed’ and a
    customer invoice will be created. Inventory on-hand will also automatically
    be updated.
2. Use the ’ **Shipping–Ship** ’ menu to ship a sales order fully or partially.
    Inventory on-hand will be updated accordingly. Later on you can open the
    sales order and create the customer invoice using the ’Customer Invoice’
    button at the bottom of the screen.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 152

**3.11.1.5 Using Shipping menu to ship a sales order**

In this section we shall explain how you can use the ’ **Shipping** ’ menu to ship a
sales order partially or in full.

The following screen is displayed when you click ’ **Shipping–Ship** ’. Here you
can define any criteria for the sales orders you want to process or just click the
’Continue’ button if you want to list all open sales orders.

The following screen shows all the sales orders with open quantities. You
click on a particular sales order to ship the goods listed in it.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 153

The selected sales order is displayed with the order quantities and you can
enter the quantities shipped in the ’Ship’ filed. If there are any serial numbers
associated with the shipped goods you can enter them in the serial number field.
You need to specify the correct shipping date and click on ’Done’ to finish the
transaction. In the example below we are only partially shipping this sales order.

If you open this sales order again (using the **’Order Entry–Reports–Sales
Orders** ’ menu) you will see the quantity shipped stated in the ’Ship’ column.
The shipped quantities will be updated every time you ship goods using the
’ **Shipping–Ship** ’ menu.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 154

**3.11.1.6 Creating customer invoice from a partially shipped sales order**

You can create a customer invoice from a sales order any time for the quantities
shipped so far. To do this click the ’Customer Invoice’ button on sales order.
The ’Add Customer Invoice’ screen will open up with the data from that sales
order as well as the shipped quantities as show below.
Once a customer invoice has been created from a sales order, that sales order
is closed. If there are still some open quantities in that sales order, then a new
sales order with the same number and the remaining items will be automatically
created.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 155

**3.11.1.7 Sales order reports**

The **’Order Entry–Reports–Sales Orders’** menu lists all your purchase orders.
You can check/uncheck the ’Open’ and ’Closed’ checkboxes on the search screen
before you continue. ’Closed’ sales orders are those which have been fully received
or which have been marked ’Closed’ by editing the sales order.

```
The sales order report will list all your open sales orders.
```

##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 156

#### 3.11.2 Purchase orders

Creating a purchase order is often the first step you take when you buy goods
and services from your vendor. You can:

1. Make a purchase order.
2. Receive a purchase order fully or partially using the ’ **Shipping–Receive** ’
    menu.
3. Create a vendor invoice from a partially or fully received purchase order.
4. If warehouses are enabled, you can also receive goods to a particular ware-
    house.

**3.11.2.1 Add a new purchase order**

The **’Order Entry–Purchase Order’** menu will display the following screen to
allow you to add a new purchase order.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 157

If you want to enter more information for each item, you can mark the check-
box next to the ’Description’ column and then click ’Update’. Now each detail
line will span 5 lines where you can enter lots of other information for each item
you order.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 158

**3.11.2.2 Editing a purchase order**

To edit an existing purchase order, you display a list of existing orders using the
**’Order Entry–Reports–Purchase Orders’** menu and click on the purchase
oder number to edit that particular purchase order.

**3.11.2.3 Creating an RFQ or vendor invoice from a purchase order**

When you have saved a purchase order, you can open it again and use it to create
a RFQ (request for quotation). When you have received quantities you can also
directly create a vendor invoice from the purchase order. (See below on how to
use the ’ **Shipping** ’ menu to partially receive a purchase order.)

**3.11.2.4 Receiving a purchase order**

There are two ways to receive a purchase order.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 159

1. Open the purchase order and click the ’Vendor Invoice’ button. The pur-
    chase order will be received in full, marked ’closed’ and a vendor invoice
    will be created. Inventory on-hand will automatically be updated.
2. Use ’ **Shipping–Receive** ’ menu to receive a purchase order fully or partially.
    Inventory on-hand will be updated. Later on you can open the purchase
    order and create the vendor invoice using the ’Vendor Invoice’ button at
    the bottom of the screen.

**3.11.2.5 Using Shipping menu to receive a purchase order**

In this section we shall explain how you can use the ’ **Shipping** ’ menu to receive
a purchase order partially or in full.

The following screen is displayed when you click ’ **Shipping–Receive** ’. Here
you can define criteria for the purchase orders you want to process or just click
the ’Continue’ button if you want to list all open purchase orders.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 160

The following screen shows all purchase orders with open quantities. You can
click on a particular purchase order to receive the goods listed in it.

The selected purchase order is displayed with the order quantities and you
can enter the quantities received in the ’Recd’ field. If there are serial numbers
associated with the received goods you can enter them in the serial number field.
You need to specify the correct receiving date and click on ’Done’ to finish the
transaction. In the example below we are only partially receive this purchase
order.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 161

If you open this purchase order again (using the **’Order Entry–Reports–
Purchase Orders** ’ menu) you will see the quantity received stated in the ’Recd’
column. The received quantity will be updated each time you receive goods using
the ’ **Shipping–Receive** ’ menu.

**3.11.2.6 Creating vendor invoice from a partially received purchase
order**

You can also create a vendor invoice for the quantities received so far. To do this
just click the ’Vendor Invoice’ button and the ’Add Vendor Invoice’ screen will
open up with the data from that purchase order as well as the received quantities
as show below.
Once a vendor invoice has been created from a purchase order, that purchase
order is closed. If there are still some open quantities in that purchase order,
then a new purchase order with same number and the remaining items will be
automatically created.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 162

**3.11.2.7 Purchase order reports**

The **’Order Entry–Reports–Purchase Orders’** menu shows you all your pur-
chase orders. You can check/uncheck the ’Open’ and ’Closed’ checkboxes on the
search screen before you continue. ’Closed’ purchase orders are those which have
been fully received or which have been marked ’Closed’ by editing the purchase
order.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 163

**3.11.2.8 Order entry notes**

Please note that:

1. Usually inventory on-hand quantities are updated when you create a vendor
    or customer invoice. This default behavior is changed if you are using orders
    module. See below.
2. When you receive or ship an order through the ’ **Shipping** ’ menu, your
    inventory on-hand is immediately updated. You can confirm this by viewing
    ’ **Warehouses–Reports–Onhand** ’ after you receive or ship an order. Your
    accounts receivable and accounts payable are only updated when you create
    an invoice from a partially or fully received order.
3. You cannot change the listed item quantities or add new items when an
    invoice is created from a partially or fully shipped/received order. You can
    only add services to an invoice created from a partially or fully shipped/re-
    ceived order. This feature is needed to keep the invoices and orders quan-
    tities data in sync.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 164

4. When you create an invoice from a partially shipped or received order,
    SQL-Ledger closes that order and creates a new one with the remaining
    order quantities but with same order number.

#### 3.11.3 Important inventory on-hand reports

1. Inventory on hand at warehouses: The **’Warehouses–Reports–Onhand’**
    report. See 2.13.6 for details.
2. Inventory receive/ship activity: The **’Warehouses–Reports–Activity’** re-
    port. See 2.13.7 for details.

### 3.12 Time Cards

Time cards module allows you to record the time you have spent to provide a
service to your customer. The work flow for using time cards goes like this:

1. Create a project for the customer.
2. Create time card entries.
3. Create a sales order.

We go through each of these steps using screen shots below.

#### 3.12.1 Create a project for the customer

You can create a new project using the ’ **Projects–Add Project** ’ menu. Here
you can also insert the name of the customer for whom you or your staff will be
working. You can also specify start- and end dates as desired.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 165

#### 3.12.2 Create time card entries

Once you have created the project for the customer, you can start creating time
card entries. Use the ’ **Projects–Add Time Card** ’ menu to add a new time card
entry. On this screen you need to select the employee name, project, date and
then specify the time worked. You also need to select the service code (article
number) of the service you provided, as you defined it using the ’ **Goods &
Services–Add Service** ’ menu. If you are not sure about the article number for
a certain service, just enter ’%’ and click ’Update’. This will produce a list of all
your available services and their respective article number.

Once you have added time cards, you can view a report for the selected time
cards using the ’ **Projects–Reports–Time Cards** ’ menu.

```
By clicking on the ’ID’ link in the list, you can edit the time card.
```

##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 166

#### 3.12.3 Create a sales order for open time cards

The third step in using time cards is to create a sales order for your open time
cards. To do this you use the ’ **Projects–Generate–Sales Orders** ’ menu and
select the projects for which you want to create a sales order.

You select the required project and click on the ’ **Generate Sales Orders** ’
button to create your sales order. Once the sales order has been generated, you
can view it using ’ **Order Entry–Reports–Sales Orders** ’ menu.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 167

### 3.13 Audit Control

You can use the ’ **System–Audit Control** ’ menu to enforce transaction control
and log user activities.

#### 3.13.1 Enforce transaction reversal for all dates

You can check this option to prevent any change to posted transactions. If the
’Enforce transaction reversal for all dates’ checkbox has been marked, you need
to add a new reverse transaction each time you want to correct some mistake.
This option is highly recommended to keep your transactions fully accountable.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 168

#### 3.13.2 Close books up to

When you close your books up to a certain date, SQL-Ledger will not allow the
changing of any transaction that has been booked prior to that date. Please note
that this is not a year end process, but merely a precaution to prevent changes
in periods that have been reconciled.

#### 3.13.3 Activate audit trail

When you mark the ’Activate audit trail’ checkbox, all user activities (adding,
changing and deleting transactions) are logged. You can view the log by using
the **’System–Audit Trail’** report.

#### 3.13.4 Remove audit trail up to

You can use this option to remove the audit trail from your database up to a
certain date. Please note that according to the legislation in some countries
you may need to be able to provide an audit trail for the last ten years of your
accounting.

### 3.14 Reconciliation

The account reconciliation function in the ’ **Cash** ’ module allows you to match
your SQL-Ledger transactions with, for example, your bank statement and then
mark them as reconciled. This way you can make sure that your account balance
in the bank matches your account balance in SQL-Ledger up to a certain date.

#### 3.14.1 Marking transactions

To match and mark your account transactions in SQL-Ledger with, for example,
your bank statement you need to open the reconciliation screen using the ’ **Cash–
Reconciliation** ’ menu. Here you first select the account you want to reconcile,
enter the period and click ’Continue’. Please note the ’Usage Notes’ section on
this screen, which will help you display the account transactions as you desire.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 169

Once you have displayed your account transactions, you can check/uncheck
the checkbox which is next to the description column. If the transaction is
reconciled, check this box, if not then don’t. Once you have checked all the
reconciled transactions you can click the ’Update’ followed by the ’Done’ button
to save the updates.

The reconciliation report allows you to view your reconciled account with its
balance.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 170

### 3.15 Year end

The **’System–Yearend’** menu creates a general ledger transaction which clears
the income and expense accounts in SQL-Ledger and posts the difference (which
is income or loss) to the specified retained earnings account.
Please note that:

1. The year-end process can be run daily, weekly, monthly, quarterly or yearly.
2. The year-end general ledger transaction is not included in the income state-
    ment which covers the period containing a closing transaction.
3. The year-end general ledger transaction can be viewed through the ’ **General**
    **Ledger–Reports** ’ menu and edited or deleted as required.
4. The year-end process does not automatically close your books. Please see
    3.13.2 above for information on how to close your books up to a certain
    date.

This is the year end screen followed by an example of the general ledger entry
created during the year-end process.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 171

### 3.16 Data backup

You can backup your database using the ’ **System–Backup** ’ menu. There are
two ways to get your backup:

#### 3.16.1 Send by Email

When you click this menu option, the backup is sent to your email address by
email. You can add or change your email address in the ’ **Preferences** ’ menu.

#### 3.16.2 Save to File

When you click this menu option, your browser will display the save file dialog
and you can save the backup file on your local computer.

### 3.17 Basics of double entry accounting

#### 3.17.1 Introduction

The double entry accounting system, although many times feared by non-accountants,
is a very simple but extremely powerful method of managing money. SQL-Ledger


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 172

does much of the double entry accounting itself linking all parts of the application
through the chart of accounts.
You only need to know how the double entry accounting system works when
you are going to make general ledger transactions. Its basic principle is that every
business transaction affects at least two accounts. For example:

- When you buy a car, you cash is decreased and your assets are increased.
- When you sell a item on cash, your sale is increased and your cash is also
    increased.

#### 3.17.2 Account types

There are five basic types of accounts which are given below:

1. Assets
2. Liabilities
3. Equity
4. Sales
5. Expenses

#### 3.17.3 Accounting rules

- Assets (1) and Expenses (5) are increased by debit and decreased by credit.
- Liabilities (2), Equity (3) and Sales (4) are increased by credit and de-
    creased by debit.

#### 3.17.4 Examples

**You invest $1000 to start a new business:**

- Debit: Your bank account
- Credit: Equity account


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 173

**You pay $100 check for office rent:**

- Debit: Office rent expense accoun
- Credit: Your bank account

**You build a website for a customer asking him to pay $200. Customer
promises to pay after 20 days.**

- Debit: Accounts Receivables (Debtors)
- Credit: Sales

**Your customer pays you $200 after 20 days.**

- Debit: Your bank account
- Credit: Accounts Receivables (Debtors)

Here is a really simple and useful accounting tutorial: [http://www.a-systems.net/accounting.htm](http://www.a-systems.net/accounting.htm)

### 3.18 Cost of goods sold (COGS)

Cost of goods sold (COGS) is the purchase price of any goods you sold. Your
sales minus the COGS is your gross profit. COGS is an important accounting
information. Correct COGS gives you a clear picture of the profitability of your
sales.
Tip: To view the debit and credit accounting transactions for any sale or
purchase invoice, enter the invoice number in the **’General Ledger–Reports’**
search screen and click the ’Continue’ button.

#### 3.18.1 Sale invoices and COGS

Let us make it clear with an example:
You purchase 10 Linux computers for $400 each.

- Debit: Inventory $4000
- Credit: AP $4000


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 174

A customer comes in and purchases 2 of these at $500 each.

- Debit: AR $1000 Credit: Sales $1000
- Debit: COGS $800 Credit: Inventory $800

So your gross profit is $200.
SQL-Ledger posts COGS automatically with each sales invoice. It calculates
COGS based on the First-In First-Out (FIFO) principle. This means is that
if you purchase 5 more Linux computers at $430 each, SQL-Ledeger will keep
calculating COGS @ $400 each until all 10 Linux computers of the first purchase
transaction are depleted. Afterward it will calculate COGS @ $430.

#### 3.18.2 Sales before purchases

SQL-Ledger allows you to sell goods without purchasing them in advance. This
is a common practice in many businesses where you have received the goods, but
do not yet have the vendor invoice.
This will results in a negative stock quantity in the ’ **Goods & Services–
Reports–All Items** ’ report. No COGS is posted for such transactions at the
time of sale. Later when you record purchases, COGS is automatically recorded
for these oversold items.

#### 3.18.3 Editing sale invoices

When you edit and repost an already posted invoice, COGS goes out of sync and
incorrect accounting entries are posted. This causes incorrect income statement.
To confirm this, display your income statement and write down the COGS
amount. Now open and repost any past sales invoice. Compare the new COGS
in income statement with the old one.
Ideally you should never edit an invoice. Instead post a reversal of the in-
voice (using a credit invoice) and create a new invoice. Check the box ’Enforce
transaction reversal for all dates’ on **’System–Audit Control’** screen.
You can correct the incorrect COGS which was booked when you edited and
reposted an already posted invoice transaction, by running the re-posting script
in the ’ **System–Maintenance–Repost Invoices** ’ menu.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 175

### 3.19 Ledger Doctor

Ledger Doctor is a tool to identify database inconsistencies in the SQL-Ledger
database. Use the **’System–Maintenance–Ledger Doctor’** menu to access
it. The ’ **Ledger Doctor** ’ report itself does not correct any error, it only reports
inconsistencies with hints on how to correct them.

### 3.20 Monitor

Using the **’System–Maintenance–Monitor’** menu, you can run any individual
SQL query or command directly on your SQL-Ledger database.


##### CHAPTER 3. RUNNING YOUR BUSINESS ON SQL-LEDGER 176

```
WARNING: Be careful with this option as no checks are made on what you do.
You can quickly corrupt your database with a small mistake. If you are not sure
how to use it then just ignore it.
```
```
TIP: Always take a backup before running any SQL using this menu.
```

# Chapter 4

# Keeping track of your business in

# SQL-Ledger

This sections explains the various reports which are available in SQL-Ledger to
monitor and track your business once you have started recording your business
transactions. SQL-Ledger stores all your business data in an SQL database. SQL,
which stands for Structured Query Language, is a special purpose programming
language designed for managing data held in a relation database management
system. SQL is also a standard of the International Organization for Standard-
ization (ISO).
Running SQL queries on a business database can be a very complex matter
and usually requires basic knowledge regarding the individual database structure.
The developers of SQL-Ledger made it one of their major goals to simplify this
process of SQL queries and were able to find a unique way to make it an easy
task for anyone to analyze the business data stored inside the database, even
without knowledge in SQL.
There are many different reports in SQL-Ledger and they can all be divided
into two main groups:

- a.) **Financial Reports** , which reflect the financial effects of your business
    transactions and
- b.) **Module Reports** , which enable you to analyze the various details
    behind your business transactions.

The Financial Reports are listed in the menu under ’ **Financial Reports** ’ and
the module reports are listed under ’ **Reports** ’ in the menu of each individual

##### 177


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 178

module. For example, the Accounts Receivable module reports are listed in the
menu under ’ **AR–Reports’** and the Goods & Services module reports are listed
in the menu under **’Goods & Services–Reports** ’.

### 4.1 Financial reports

There are seven different financial reports in the ’ **Financial Reports** ’ menu:
Chart of accounts, trial balance, income statement, balance sheet, tax report,
project income statement and department income statement.

#### 4.1.1 Chart of accounts & trial balance

The chart of accounts report and the trial balance report are both standard
accounting reports which show amounts posted to each individual account in your
chart of accounts. They show all transactions posted on the individual accounts
from all modules. The chart of accounts report shows the total amounts booked
in debit and credit, whereas the trial balance report also shows the beginning
balance and ending balance of each single account.
In the chart of accounts report you first choose the individual account and
then set the period to be shown. In the trial balance report you start by choosing
the period and then select the individual account.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 179

#### 4.1.2 Income statement

The income statement is a financial report that lists income, expenses and profit
or loss for a given period of time. Income statements can be run for any period
and you can also compare the results with previous periods. The structure and
presentation of your income statement can be changed to suit your individual
needs, either by linking each individual account to a GIFI account, or by including
headers in your chart of accounts. See 2.9.5 for more information on how to setup
GIFI accounts and account headings.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 180

```
The following income statement is for a single period.
```
```
The following income statement includes two periods.
```
#### 4.1.3 Balance sheet

The balance sheet is a financial statement that lists the assets, liabilities, and
the ownership equity of a business entity as of a specific date. The balance sheet
can be displayed as of any particular date. Like the income statement, you can
also compare it with the account totals of previous dates.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 181

#### 4.1.4 Tax report

The tax report is a consolidated statement of all taxable and non-taxable accounts
payable (AP) and accounts receivable (AR) transactions. Tax reports can be
shown and printed for a certain month, quarter, year or any other defined period
of time. At the top of the tax report you can find the consolidated totals for each
account and below that you can see the individual accounts, single transactions
and their totals.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 182

#### 4.1.5 Project & department income statement

The project income statement lists income, expenses and profit or loss for se-
lected projects, and the department income statement does the same for selected
departments.

#### 4.1.6 Project Income statement

On search screen you can select which projects you want to include in the report.

```
Click ’Continue’ button to view the report for all or selected projects.
```

##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 183

#### 4.1.7 Department Income statement

On search screen you can select which departments you want to shown income
statement.

```
Click ’Continue’ button to view the report for all or selected departments.
```

##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 184

### 4.2 Module reports

All module reports in SQL-Ledger have been preconditioned to display the basic
information usually required when analyzing data in that module. For example,
the preconditioned module report in ’ **AR–Reports–Transactions** ’ will automat-
ically display the “Date, Invoice Number, Description, Customer, Total and Paid
Amounts” of the open account receivables.
One of the major strengths of SQL-Ledger is that all module reports can easily
be customized to fit individual needs or requirements. To adapt a module report
to your individual requirements, all you need to do is to enter criteria, select
report columns with check boxes and click the ’Continue’ button to display the
report.

#### 4.2.1 AR reports

There are six main AR reports in SQL-Ledger; transactions, outstanding, AR
aging, reminder, tax collected and non-taxable. You will find all these reports in
the menu under **’AR–Reports’**.
The first thing you will see when you select one of these reports is the search
screen. In the search screen you can enter different criteria for your report and
also select which specific data you want to display in the report.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 185

All reports are preset to display the information which is most commonly
required, so if you simply click on the ’Continue’ button without defining any
criteria or selecting specific data, the standard report will be displayed.

**4.2.1.1 Transactions report**

Transaction report shows all currently open or closed transactions and invoices
for the specified criteria on search screen. On search screen you can specify
various criteria and select/ de-select columns which you want to shown on the
screen.

When you click ’Continue’ button after specifying the required criteria, your
report is displayed. This is ’Summary’ report where each invoice or transaction
is shown on single line.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 186

The ’Detail’ report shows debits and credits of each transaction along with
account number. In detail report, a single invoice or transaction is shown on
multiple times. You can click ’Subtotal’ to subtotal and group this report by
invoice number.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 187

**4.2.1.2 Outstanding report**

The difference between the outstanding report and the transactions report, is that
the latter will show you only the selected open or closed invoices and transactions
as of today.
The outstanding report, on the other hand, will show you the selected open
and closed invoices and transactions as on a chosen date or as within a chosen
time frame. If you for example want to create a report for an auditor to show
which invoices and transactions were still open on the 31. December of last year,


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 188

you need to use the outstanding report, since the transactions report only shows
the selected open invoices and transactions as of today. On the outstanding
report search screen you can specify various criteria and select/ de-select columns
which you want to shown on the screen.

**4.2.1.3 AR aging report**

AR aging report shows the outstanding balances of your customers divided into
predefined periods of time in the past.

The summary aging report (shown below) shows one line for each customer
with outstanding balance.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 189

The detail aging report (shown below) shows all outstanding invoices for each
customer with subtotal by the customer.

**4.2.1.4 Reminder report**

In the reminder report you can also print or email reminders to your customers.
The reminder report search screen can be set to show selected departments or
selected customers.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 190

If neither is selected it will display all outstanding customer invoices and divide
them by currency.

You can define up to 3 levels of reminders. Level 1 being a polite one and
level 3 being a bit harsh one. When you print a reminder, its level is stored in
the database. Next time when you run the reminder report, the next level of
reminder for that customer is displayed. You also have the option to change the
reminder level manually and click on ’Save Level’.

**4.2.1.5 Tax collected and non-taxable reports**

The **’AR–Reports–Tax collected’** and **’AR–Reports–Non-taxable’** reports
are statements of all taxable and non-taxable customers (AR) transactions.
These tax reports will display a statement with the single transactions and their
totals for a chosen month, quarter, year or any other defined period of time.

#### 4.2.2 Customers reports

There are two customer reports in SQL-Ledger; **’Customers–Search’** and ’ **Customers–
History** ’. The search report is used to display customer master data and can also


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 191

display the customer related business transactions. The history report is used to
display customer related totals for selected business transactions.
The search report is based on the total value of customer business transac-
tions, whereas the history report is based on the total quantities of customer
business transactions.

**4.2.2.1 Customer search report**

The customer search report can be used either to find and make changes to
existing customers or to list the individual business transactions for selected cus-
tomers.

If you select the ’Sales Invoice’ checkbox in the customer report search screen,
SQL-Ledger will display all sales invoices that have been issued for the selected
customer and their respective amount, tax and total values. You can also click
’Subtotal’ in the search screen to subtotal the values by customer.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 192

**4.2.2.2 Customer history report**

You can use history reports to see which customer is giving you more business.
You can filter the report on date range which is applied to the invoices (or orders
or quotations).

```
Customer history summary report shows business activity grouped by item.
```

##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 193

```
Customer history detail report shows business activity by invoice and item.
```

##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 194

#### 4.2.3 AP reports

There are five main AP reports in SQL-Ledger; transactions, outstanding, AR
aging, tax collected and non-taxable. You will find all these reports in the menu
under **’AP–Reports’**.
The first thing you will see when you select one of these reports is the search
screen. In the search screen you can enter different criteria for your report and
also select which specific data you want to display in the report.
All reports are preset to display the information which is most commonly
required, so if you simply click on the ’Continue’ button without defining any
criteria or selecting specific data, the standard report will be displayed.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 195

**4.2.3.1 Transactions report**

Transaction report shows all open or closed transactions and invoices for the
specified criteria on search screen. On search screen you can specify various
criteria and select/de-select columns which you want to shown on the screen.

When you click ’Continue’ button after specifying the required criteria, your
report is displayed. This is ’Summary’ report where each invoice or transaction
is shown on single line.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 196

The ’Detail’ report shows debits and credits of each transaction along with
account number. In detail report, a single invoice or transaction is shown on
multiple times. You can click ’Subtotal’ to subtotal and group this report by
invoice number.

**4.2.3.2 Outstanding report**

The difference between the outstanding report and the transactions report, is that
the latter will show you only the selected open or closed invoices and transactions
as of today.
The outstanding report, on the other hand, will show you the selected open


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 197

and closed invoices and transactions as on a chosen date or as within a chosen
time frame. If you for example want to create a report for an auditor to show
which invoices and transactions where still open on the 31. December of last
year, you need to use the outstanding report, since the transactions report only
shows the selected open invoices and transactions as of today.
On the outstanding report search screen you can specify various criteria and
select/ de-select columns which you want to shown on the screen.

**4.2.3.3 AP aging report**

AP aging report shows the outstanding balances of your customers divided into
predefined periods of time in the past.

The summary aging report (shown below) shows one line for each customer
with outstanding balance.

The detail aging report (shown below) shows all outstanding invoices for each
customer with subtotal by the customer.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 198

**4.2.3.4 Tax paid and non-taxable reports**

The ’ **AP–Reports–Tax collected** ’ and ’ **AP–Reports–Non-taxable** ’ reports
are statements of all taxable and non-taxable vendor (AP) transactions. These
tax reports will display a statement with the single transactions and their totals
for a chosen month, quarter, year or any other defined period of time.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 199

#### 4.2.4 Vendor reports

There are two vendor reports in SQL-Ledger; **’Vendors–Reports–Search’** and
’ **Vendors–Reports–History** ’.
The search report is used to display vendor master data and can also display
the vendor related business transactions. The history report is used to display
vendor related totals for selected business transactions. The search report is
based on the total value of vendor business transactions, whereas the history
report is based on the total quantities of vendor business transactions.

**4.2.4.1 Vendor search report**

The Vendor search report can be used either to find and make changes to existing
vendors or to list the individual business transactions for selected vendors.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 200

If you select the ’Vendor Invoice’ checkbox in the vendor search report search
screen, SQL-Ledger will display all purcahse invoices that have been issued for
the selected vendor and their respective amount, tax and total values. You can
also click ’Subtotal’ in the search screen to subtotal the values by vendor.

**4.2.4.2 Vendor history report**

You can use history reports to see which vendor you buy most from and which
vendor you buy less from and what. You can filter the report on date range which
is applied to the invoices (or orders or quotations).


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 201

```
Vendor history summary report shows purchase activity grouped by item.
```
```
Vendor history detail report shows purchasing activity by invoice and item.
```

##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 202

#### 4.2.5 Cash reports

There are three main cash reports in SQL-Ledger; receipts, payments and recon-
ciliation.

**4.2.5.1 Receipts**

Receipts report will list all receivables (incoming) payments that have been
booked on the available payment accounts.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 203

**4.2.5.2 Payments**

Payments report will list all payables (outgoing) payments that have been booked
on the available payment accounts.

**4.2.5.3 Reconciliation**

Reconciliation report will list all transactions that have been marked as reconciled
on any chosen account. See 3.14 to learn more about how to mark transactions
as reconciled.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 204

#### 4.2.6 Order entry reports

There are three main order entry reports in SQL-Ledger; sales orders, require-
ments and purchase orders.

**4.2.6.1 Sales orders**

In the sales order search screen you can define criteria for the purchase orders
you want to list. For example, you can check/uncheck the ’Open’ and ’Closed’
to list only open or closed sales orders.

’Closed’ sales orders are those which have been fully received or which have
been marked ’Closed’ by editing the sales order.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 205

**4.2.6.2 Requirements**

The requirements report will show you which parts and assemblies are low on
stock and need to be ordered or assembled. The reorder point (ROP) is set
individually for each part or assembly by entering the desired minimum quantity
in the ROP field. The requirements report will show which parts and assemblies
need to be ordered or assembled based upon the following factors:
On-hand quantity
Open Sales Orders
Open Purchase Orders


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 206

**4.2.6.3 Purchase orders**

In the purchase order search screen you can define criteria for the purchase orders
you want to list. For example, you can check/uncheck the ’Open’ and ’Closed’
to list only open or closed purchase orders.

’Closed’ purchase orders are those which have been fully delivered or which
have been marked ’Closed’ by editing the purchase order.

#### 4.2.7 Warehouses reports

There are four main warehouse reports in SQL-Ledger; transfers, deliveries, on-
hand and activity.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 207

**4.2.7.1 Transfers**

You can move inventory between warehouses by using the ’ **Warehouses–Add
Transfer** ’ menu option. The transfers report will show you any inventory trans-
fers that have been done between warehouses.

**4.2.7.2 Deliveries**

Some companies need to track the in-transit goods between warehouse transfers.
The delivery date is sometimes different from the transfer date. The deliveries
report will display all the transfers pending to be received. To ’receive’ the
transfers, specify the dates when the goods were delivered at ’your’ warehouse
and click ’Save Delivered’.

#### 4.2.8 Quotations reports

There are two main quotation reports in SQL-Ledger; quotations and RFQs
(request for quotations).

**4.2.8.1 Quotations**

The quotation report will display the existing quotations. You can specify any
conditions to filter the report by entering your criteria in the search screen or just
leave it blank and click ’Continue’ button to get all existing ’Open’ quotations.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 208

’Closed’ quotations are those which have been used to create a sales order or
which have been marked ’Closed’ by editing the quotation.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 209

**4.2.8.2 RFQs**

RFQs are quotation requests that have been sent to your vendors. The RFQs
report will display the existing quotations. You can specify any conditions to filter
the report by entering your criteria in the search screen or just leave it blank and
click ’Continue’ button to get all existing ’Open’ quotations.
’Closed’ quotations are those which have been used to create a purchase
order or which have been marked ’Closed’ by editing the quotation.

#### 4.2.9 General ledger reports

The general ledger reports is used to view all accounting journals with debits and
credits to the particular accounts. Initially this report can be confusing because
it shows not only the journals added using ’Add Transaction’ menu show above
but also all accounting transactions posted from AR, AP and cash modules.
You can specify any conditions to filter the report by entering your criteria
in the search screen or just leave it blank and click ’Continue’ button to list all
existing general ledger transactions.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 210

The General ledger reports can also be used to export all or certain defined
transactions in CSV-format. To achieve this, just mark the CSV checkbox before
you click on ’Continue’.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 211

```
General ledger reports can be sorted on any displayed column.
```
#### 4.2.10 Project reports

There are three main project reports in SQL-Ledger; projects, transactions and
time cards.

**4.2.10.1 Projects**

The projects reports will display a list of all available projects and their respective
start and end dates.


##### CHAPTER 4. KEEPING TRACK OF YOUR BUSINESS IN SQL-LEDGER 212

**4.2.10.2 Transactions**

The transactions report will display the the total amounts booked in debit and
credit on each account for any chosen project, and include the beginning balance
and ending balance of each account. By clicking on the account number you can
also drill down and see the individual transactions.

**4.2.10.3 Time cards**

The time card report will display time cards that have been entered for any chosen
project.


# Chapter 5

# Ledger Cart

### 5.1 Introduction

LedgerCart instantly creates an on-line store and order system using information
in your SQL-Ledger. You just drop the cgi scripts into your web server, install
few CPAN modules, configure your db connection and you are ready to go.
Users can browse products and services, add items to their cart and checkout
in a familiar way. New order is added to SQL-ledger sales orders.

#### 5.1.1 Features

1. Extremely simple to install and configure.
2. Can be installed on dedicated or shared hosting.
3. No additional database required. Retrieves and saves all data from/to
    SQL-Ledger dataset.
4. Easy to customize. All pages are standard html pages with template toolkit
    tokens.
5. Add new pages by creating standard html files and linking them in header.html
    or sidebar.html.
6. Look and feel can be customized using css and templates.
7. A single script ’index.pl’ allows you to easily add more features by adding
    new actions.

##### 213


##### CHAPTER 5. LEDGER CART 214

8. Add item descriptions. These are displayed on product detail page and are
    stored in item notes. Item descriptions can use markdown syntax.
9. Add item images. LedgerCart automatically creates thumbnails and shows
    full image on item detail.
10. Visitors can now add items to their cart and checkout with their billing and
shipping address.
11. New customers can register during checkout.
12. Existing customers can get a new password to their email using ’forgot
password’. They can login with their email address and place orders.
13. Customers can browse their orders and invoices when logged-in.

#### 5.1.2 Limitations

Currently no payment gateways are supported. However, you can sponsor the
development of any payment gateway you want to use with LedgerCart. Send
email to support@ledger123.com for details.

#### 5.1.3 Using LedgerCart as an online store

LedgerCart can instantly turn your SL installation into an on-line store with little
or no effort. Customers can place order using the familiar shopping cart interface.
Your existing customers can generate a new password using ’Forgot password’
feature.

#### 5.1.4 Using LedgerCart as Self service portal

LedgerCart can be used to serve as a self-service internet portal just like the
self-service internet banking. Your customers can view:

1. Their orders summary, order details and status
2. Invoices summary and details
3. Statements (payment summary and detail)


##### CHAPTER 5. LEDGER CART 215

#### 5.1.5 Screen shots

Here are some screen shots.


##### CHAPTER 5. LEDGER CART 216


##### CHAPTER 5. LEDGER CART 217

### 5.2 Installation

#### 5.2.1 Software packages

Login to the server with your user name and password. To be able to install
the software, we have to change to the “root” account. In this way, we get
administrator rights. Type:
su -

```
and enter your password.
With the following command, we install the packages we need for LedgerCart:
```
apt-get install libcgi-simple-perl libdbi-perl libtemplate-perl libobject-
signature-perl libnumber-format-perl libmime-lite-perl libdbix-simple-
perl libtext-markdown-perl libdate-calc-perl libgd-gd2-perl
libdatetime-perl libhtml-format-perl apg
After that you need to install some further cpan modules:

cpan GD cpan GD::Thumbnail cpan MIME::Lite::TT::HTML

```
Then install LedgerCart in your SQL-Ledger directory:
```
git clone git://github.com/ledger123/ledgercart.git ledgercart

#### 5.2.2 Configuration and Admin access

To configure LedgerCart for your installation, edit the config.pl file and change
the appropriate lines for your database connection information. You can also
change default thumbnail sizes here.


##### CHAPTER 5. LEDGER CART 218

**5.2.2.1 Admin User**

To enable admin access, create a customer using SQL-Ledger with your email
address and specify its id in $form{admin_id}. Now using “forgot password”
link, generate a new password which will be sent to your email address.

**5.2.2.2 Editing item descriptions, images and thumbnails**

When you are logged in as admin and visit item detail pages, you can edit item
descriptions as well as upload images and auto-create thumbnails.
Item descriptions text uses simple markup language ’markdown’ for html ele-
ments. No html is allowed for security reasons. See [http://daringfireball.net/projects/markdown/dingus](http://daringfireball.net/projects/markdown/dingus)
for markdown syntax. Item descriptions are stored in item notes column and can
be editing from within SQL-Ledger as well.

**5.2.2.3 Editing pages through admin access**

Once you login as admin, you can see ’Edit’ links. Pages can be edited right
away. You can use standard html and template toolkit tokens to edit pages.

**5.2.2.4 Marking ’hot’ and ’new’ items**

When you are logged in as admin, add items to your cart and click the ’Save
cart as hot items’ or ’Save cart as new items’. This will mark those items as
hot or new and will display them on man page (in default templates). In future,
hot/new functionality will be made to work based upon actual ’hot’ or ’new’
items.

#### 5.2.3 Customization

LedgerCart is extremely easy to customize. LedgerCart consists of one big
gateway script ’index.pl’ which processes html templates created with Tem-
plate::Toolkit.

1. Template::Toolkit templates are standard html files which can include Perl
    variables within [% and %] delimiters. You can copy the default templates
    and modify them as you please.
2. New pages can be added by creating standard html files and linking them
    to ’templatesfolder/header.html’ or ’templatesfolder/sidebar.html’.


##### CHAPTER 5. LEDGER CART 219

3. You can also customize the theme.css to change the colors and other look
    and feel according to your taste.
4. Expert users can modify the ’index.pl’ file to add their own variables which
    can be interpolated within your LedgerCart templates.


# Chapter 6

# Development and Customization

### 6.1 Customization

SQL-Ledger can be customized in three ways:

#### 6.1.1 custom_xx.pl files

You can create your own functions or override any existing function by creating
custom scripts in custom_xx.pl files and putting them in bin/mozilla folder. For
example, to add new functions to gl.pl file, add these functions to custom_gl.pl
file and put this file into bin/mozilla/ folder. This file will be automatically loaded
by SQL-Ledger before running any functions in gl.pl files.
Once your new functions are there, you can call them using your own custom
menu. Custom menu entries are put in custom_menu.ini and follow the same
syntax as that of menu.ini. This method of extending the SQL-Ledger is upgrade-
safe and is the recommended way.

**6.1.1.1 Custom Modules**

You can build your own modules. To write a module, you need to create at least
three files:

1. Module back-end code which will reside in ./SQL-Ledger/SL/MyModule.pm
2. Module front-end code which will reside in ./SQL-Ledger/bin/mozilla/mymodule.pl

##### 220


##### CHAPTER 6. DEVELOPMENT AND CUSTOMIZATION 221

3. Gateway script in ./SQL-Ledger. (You just need to make a copy of an
    existing one. For example cp gl.pl mymodule.pl in ./SQL-Ledger/ folder.

This method is also upgrade safe.

#### 6.1.2 Modify the source code

Sometimes there is a need to directly alter the SQL-Ledger source code for
particular needs. We have, for example, modified few reports (GL Transactions,
All Items) in this way. Your changes, however, will be overwritten when you
upgrade to new version and you will need to port these changes again to the new
version.
A bit discipline and an SCM software like GIT can help manage such changes
or patches with easy. We, at ledger123.com, use GIT to track and manage such
changes across newer versions of SQL-Ledger.

### 6.2 Adding a new translation

SQL-Ledger can be run in 45 languages. Each user in a single installation can
run it in his or her own language. So, for example, users of a company with
offices in Germany, Italy and France, can see SQL-Ledger in their own native
languages.
If your language is missing then you can add it using language translation
feature. This feature can also be used to customize the user interface text to suit
your business needs. For example you can translate ’Customers’ to ’Students’ if
you are using SQL-Ledger in a school or college accounting department.
Here the steps you need to take to create a new language:

1. Add a new folder in within ’locale’ folder.
2. Create a text file LANGUAGE with short description of your language.
3. Copy locales.pl from ’locale/de/’ folder to this folder.
4. Create a text file named ’all’ with following format. Here you translate the
    default English labels to any text in English or your native language:


##### CHAPTER 6. DEVELOPMENT AND CUSTOMIZATION 222

```
$self{texts} = {
’Shipping Point’ => ’Shipping Place’,
’Ship via’ => ’Ship Name’,
’Waybill’ => ’Bill Number’,
}
```
5. One you have created this ’all’ file with all the required strings for your
    translations, you will run ’perl locales.pl -m’ on command line and the
    translation files for all modules will be created individually.

Please note that this ’all’ file serves as a default for all translation files which are
created by running ’perl locales.pl -m’. You can fine tune each module translation
by editing that file directly in the text editor and adding the module specific
translation in the same format as for ’all’ file.

### 6.3 SQL Queries

These sql queries for SQL-Ledger can be used in phpPgAdmin or psql.

#### 6.3.1 Simple SQL Queries

**6.3.1.1 Sales summary report**

**SELECT**
ar.invnumber,
ar.transdate,
c.name **AS** customer,
ar.netamount,
ar.amount - ar.netamount **AS** tax,
ar.amount,
ar.paid,
ar.invoice
**FROM** ar
**JOIN** customer c **ON** (c.id = ar.customer_id);

**6.3.1.2 Sales summary report with department and warehouse**

SELECT
ar.invnumber,
ar.transdate,
c.name AS customer,
ar.netamount,


##### CHAPTER 6. DEVELOPMENT AND CUSTOMIZATION 223

ar.amount - ar.netamount AS tax,
ar.amount,
ar.paid,
ar.invoice,
d.description AS department,
w.description AS warehouse
FROM ar
JOIN customer c ON (c.id = ar.customer_id)
JOIN department d ON (d.id = ar.department_id)
JOIN warehouse W ON (w.id = ar.warehouse_id);

**6.3.1.3 Sales report with items**

SELECT
ar.invnumber,
ar.transdate,
c.name AS customer
p.partnumber,
ar.description,
i.qty,
i.sellprice,
i.qty * i.sellprice AS extended
FROM ar
JOIN customer c ON (c.id = ar.customer_id)
JOIN invoice i ON (i.id = ar.trans_id);

**6.3.1.4 List of customers**

SELECT
customernumber,
name,
creditlimit
FROM customer
WHERE LOWER(name) LIKE ’%bank%’
ORDER BY name;

**6.3.1.5 Cash accounts with current balances**

SELECT
accno,


##### CHAPTER 6. DEVELOPMENT AND CUSTOMIZATION 224

description,
(
SELECT SUM(amount) FROM acc_trans
WHERE acc_trans.chart_id = chart.id
) AS balance
FROM chart
WHERE link LIKE ’%_paid%’;

**6.3.1.6 Parts list**

SELECT
p.partnumber,
pg.partsgroup,
p.description,
p.lastcost,
p.rop,
p.rop * p.lastcost AS reorder_amount
FROM parts p
JOIN partsgroup pg ON (pg.id = p.partsgroup_id)
WHERE inventory_accno_id IS NOT NULL
ORDER BY partnumber;

#### 6.3.2 Advanced SQL Queries

**6.3.2.1 Inventory on hand on specific date**

SELECT
p.partnumber,
p.description,
pg.partsgroup,
p.unit,
(
SELECT SUM(0-i.qty) AS onhand
FROM invoice i
JOIN ap ON (ap.id = i.trans_id)
WHERE ap.transdate <= ’01-01-08’ AND i.parts_id = p.id
) AS purchase,
(
SELECT SUM(i.qty) AS onhand


##### CHAPTER 6. DEVELOPMENT AND CUSTOMIZATION 225

FROM invoice i
JOIN ar ON (ar.id = i.trans_id)
WHERE ar.transdate <= ’01-01-08’
AND i.parts_id = p.id
) AS sale
FROM parts p
LEFT JOIN partsgroup pg
ON (pg.id = p.partsgroup_id);

**6.3.2.2 Customer balances on a specific date**

SELECT
ct.id,
ct.customernumber,
ct.name,
SUM(0 - ac.amount) AS balance
FROM customer ct
JOIN ar aa ON (ct.id = aa.customer_id)
JOIN acc_trans ac ON (aa.id = ac.trans_id)
JOIN chart c ON (c.id = ac.chart_id)
WHERE (ac.transdate <= ’06-30-2007’)
AND (c.link = ’AR’)
GROUP BY 1,2,3
ORDER BY customernumber;

**6.3.2.3 Sales summary by month**

SELECT
TO_CHAR(transdate, ’YY-MM’) AS month,
d.description AS department,
SUM(netamount)
FROM ar
JOIN department d ON (d.id = ar.department_id)
WHERE (transdate BETWEEN ’01.07.2005’ AND ’30.06.2006’)
GROUP BY TO_CHAR(transdate, ’YY-MM’), d.description;

**6.3.2.4 Sales Summary by group and month**

SELECT
d.description AS department,
pg.partsgroup,


##### CHAPTER 6. DEVELOPMENT AND CUSTOMIZATION 226

TO_CHAR(ar.transdate, ’YY-MM’) AS month,
SUM(0 - i.qty * i.sellprice) AS amount
FROM invoice i
JOIN ar ON (ar.id = i.trans_id)
JOIN parts p ON (p.id = i.parts_id)
JOIN partsgroup pg ON (pg.id = p.partsgroup_id)
JOIN department d ON (d.id = ar.department_id)
WHERE ar.transdate BETWEEN ’01.07.2005’ AND ’30.06.2006’
GROUP BY
d.description,
pg.partsgroup,
TO_CHAR(ar.transdate, ’YY-MM’)
ORDER BY 1, 2

**6.3.2.5 Cash received today with age of AR in days**

SELECT
c.accno,
c.description AS acc_title,
d.description AS department,
a.invnumber,
ct.name,
ac.transdate - a.transdate AS days,
ac.source,
ac.amount,
e.name AS salesper,
a.notes,
ac.memo
FROM ar a
JOIN acc_trans ac ON (a.id = ac.trans_id)
JOIN chart c ON (ac.chart_id = c.id)
JOIN customer ct ON (a.customer_id = ct.id)
JOIN employee e ON (a.employee_id = e.id)
LEFT JOIN department d ON (d.id = a.department_id)
WHERE (ac.transdate = ’30.05.06’)
AND(c.link LIKE ’%AR_paid%’)
AND (
a.department_id IN
(SELECT id
FROM department
WHERE description IN (’LC’,’LS’))
)
ORDER BY days;

**6.3.2.6 Trial Balance with Month Headings**

SELECT
accno,
description,
(SELECT SUM(amount) FROM acc_trans ac WHERE ac.chart_id = chart.id AND TO_CHAR(
transdate, ’YY-MM’) = ’06-01’) AS jan,
(SELECT SUM(amount) FROM acc_trans ac WHERE ac.chart_id = chart.id AND TO_CHAR(
transdate, ’YY-MM’) = ’06-02’) AS fab,
(SELECT SUM(amount) FROM acc_trans ac WHERE ac.chart_id = chart.id AND TO_CHAR(
transdate, ’YY-MM’) = ’06-03’) AS mar,
(SELECT SUM(amount) FROM acc_trans ac WHERE ac.chart_id = chart.id AND TO_CHAR(
transdate, ’YY-MM’) = ’06-04’) AS apr,


##### CHAPTER 6. DEVELOPMENT AND CUSTOMIZATION 227

(SELECT SUM(amount) FROM acc_trans ac WHERE ac.chart_id = chart.id AND TO_CHAR(
transdate, ’YY-MM’) = ’06-05’) AS may,
(SELECT SUM(amount) FROM acc_trans ac WHERE ac.chart_id = chart.id AND TO_CHAR(
transdate, ’YY-MM’) = ’06-06’) AS jun,
(SELECT SUM(amount) FROM acc_trans ac WHERE ac.chart_id = chart.id AND TO_CHAR(
transdate, ’YY-MM’) = ’05-07’) AS jul,
(SELECT SUM(amount) FROM acc_trans ac WHERE ac.chart_id = chart.id AND TO_CHAR(
transdate, ’YY-MM’) = ’05-08’) AS aug,
(SELECT SUM(amount) FROM acc_trans ac WHERE ac.chart_id = chart.id AND TO_CHAR(
transdate, ’YY-MM’) = ’05-09’) AS sep,
(SELECT SUM(amount) FROM acc_trans ac WHERE ac.chart_id = chart.id AND TO_CHAR(
transdate, ’YY-MM’) = ’05-10’) AS oct,
(SELECT SUM(amount) FROM acc_trans ac WHERE ac.chart_id = chart.id AND TO_CHAR(
transdate, ’YY-MM’) = ’05-11’) AS nov,
(SELECT SUM(amount) FROM acc_trans ac WHERE ac.chart_id = chart.id AND TO_CHAR(
transdate, ’YY-MM’) = ’05-12’) AS dec,
FROM chart
WHERE charttype = ’A’
ORDER BY accno;

#### 6.3.3 Queries to troubleshoot database problems

**6.3.3.1 Transactions without departments**

SELECT ’AR’, id, invnumber AS reference, transdate
FROM ar
WHERE id NOT IN (SELECT DISTINCT trans_id FROM dpt_trans)
UNION ALL
SELECT ’AP’, id, invnumber AS reference, transdate
FROM ap
WHERE id NOT IN (SELECT DISTINCT trans_id FROM dpt_trans)
UNION ALL
SELECT ’GL’, id, reference, transdate
FROM gl
WHERE id NOT IN (SELECT DISTINCT trans_id FROM dpt_trans);

**6.3.3.2 Unbalanced Journals**

SELECT ’GL’ AS mod, gl.reference, SUM(ac.amount)
FROM acc_trans ac
JOIN gl ON (gl.id = ac.trans_id)
GROUP BY 1, 2
HAVING SUM(ac.amount) <> 0
UNION ALL
SELECT ’AR’ AS mod, ar.invnumber, SUM(ac.amount)
FROM acc_trans ac JOIN ar ON (ar.id = ac.trans_id)


##### CHAPTER 6. DEVELOPMENT AND CUSTOMIZATION 228

##### GROUP BY 1, 2

HAVING SUM(ac.amount) <> 0
UNION ALL
SELECT ’AP’ AS mod, ap.invnumber, SUM(ac.amount)
FROM acc_trans ac
JOIN ap ON (ap.id = ac.trans_id)
GROUP BY 1, 2 HAVING SUM(ac.amount) <> 0
ORDER BY 3

**6.3.3.3 Orphan Transactions**

##### SELECT *

FROM acc_trans
WHERE trans_id NOT IN (
SELECT id FROM ar UNION ALL SELECT id FROM ap UNION ALL SELECT id FROM gl
);

**6.3.3.4 Correcting Assemblies Onhand**

Due to a bug/gotcha in orders handling in official SQL-Ledger, parts on hand
can go out of sync from actual transactions. Following query will help you find
the correct on hand quantity for a given assembly.

SELECT ’Purchased’, SUM(0-qty) FROM invoice WHERE parts_id = (SELECT id FROM parts WHERE
partnumber=’TW01’) AND trans_id IN (SELECT id FROM ap)
UNION ALL
SELECT ’Sold’, SUM(0-qty) FROM invoice WHERE parts_id IN (SELECT aid FROM assembly WHERE
parts_id = (SELECT id FROM parts WHERE partnumber=’TW01’)) AND trans_id IN (SELECT id
FROM ar)
UNION ALL
SELECT ’Onhand’, SUM(0-onhand) FROM parts WHERE id IN (SELECT aid FROM assembly WHERE
parts_id = (SELECT id FROM parts WHERE partnumber=’TW01’));

### 6.4 API

### 6.4.1 Introduction

SQL-Ledger allows you to call any of its functions from command line. An
example will better illustrate this.
The following code run from your Linux/Unix shell will add a new customer
to the customers table:

./ct.pl "


##### CHAPTER 6. DEVELOPMENT AND CUSTOMIZATION 229

login=armaghan
&password=armaghan
&path=bin/mozilla
&db=customer
&action=save
&typeofcontact=company
&name=Ledger123
&firstname=Armaghan
&lastname=Saqib
&city=London
"

You could also insert this information using plain old SQL INSERT statement
but here is the problem. Customer information is stored in at least three tables
(customer, contact, address). You have to make sure you INSERT rows with
correct id numbers in all three tables.
On the other hand API takes care of adding proper data rows in each tables
with a single call like above. API also validates your data and runs any logic which
is run when you are adding a customer through web interface. For example if
you have defined a sequence for customer numbers, the next number is assigned
automatically from that sequence.

### 6.4.2 API Uses

API can be used to “simulate” any SQL-Ledger function from command line. You
can add customers, vendors, parts as well as any type of transaction (invoices,
cash receipts and payments etc.)
This makes it very easy to integrate SQL-Ledger with any other application.
For example you can integrate it with your CRM solution, POS system, or e-
commerce solutions like AgoraCart or Interchange.
API also allows you to add new data entry interfaces with ease. All you need
to develop is the code which will interact with users and leave the rest to the
API.
Import invoices and payment functions built in new versions of SQL-Ledger
are in fact “newer interfaces” built using the API.


##### CHAPTER 6. DEVELOPMENT AND CUSTOMIZATION 230

### 6.4.3 Calling from PHP

You can make API calls from any language using its shell execution mechnisim.
For example you can use the following php code to make SL api call.

<?php
$module = ’./ct.pl’;
$params = ’login=armaghan’;
$params .= ’&password=armaghan’;
$params .= ’&path=bin/mozilla’;
$params .= ’&db=customer’;
$params .= ’&action=save’;
$params .= ’&typeofcontact=company’;
$params .= ’&name=Ledger123’;
$params .= ’&firstname=Armaghan’;
$params .= ’&lastname=Saqib’;
$params .= ’&city=London’;
$output = shell_exec("$module \"$params\"");
echo "<pre>$output</pre>";
?>

```
END
```


