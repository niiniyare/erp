

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;


CREATE EXTENSION IF NOT EXISTS fuzzystrmatch WITH SCHEMA public;



CREATE TYPE entity_type AS ENUM (
    'COMPANY',
    'REGIONAL',
    'DEPARTMENT',
    'COST_CENTER'
);



CREATE TYPE financial_year_status_enum AS ENUM (
    'OPEN',
    'CURRENT_FINANCIAL_YEAR',
    'CLOSED'
);



CREATE FUNCTION avgcost(integer) RETURNS double precision
    LANGUAGE plpgsql
    AS $_$

DECLARE

v_cost float;
v_qty float;
v_parts_id alias for $1;

BEGIN

  SELECT INTO v_cost, v_qty SUM(i.sellprice * i.qty), SUM(i.qty)
  FROM invoice i
  JOIN ap a ON (a.id = i.trans_id)
  WHERE i.parts_id = v_parts_id;

  IF v_cost IS NULL THEN
    v_cost := 0;
  END IF;

  IF NOT v_qty IS NULL THEN
    IF v_qty = 0 THEN
      v_cost := 0;
    ELSE
      v_cost := v_cost/v_qty;
    END IF;
  END IF;

RETURN v_cost;
END;
$_$;



CREATE FUNCTION check_department() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

declare
  dpt_id int;

begin
 
  if new.department_id = 0 then
    delete from dpt_trans where trans_id = new.id;
    return NULL;
  end if;

  select into dpt_id trans_id from dpt_trans where trans_id = new.id;
  
  if dpt_id > 0 then
    update dpt_trans set department_id = new.department_id where trans_id = dpt_id;
  else
    insert into dpt_trans (trans_id, department_id) values (new.id, new.department_id);
  end if;
return NULL;

end;
$$;



CREATE FUNCTION del_customer() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
  delete from shipto where trans_id = old.id;
  delete from customertax where customer_id = old.id;
  delete from partscustomer where customer_id = old.id;
  delete from address where trans_id = old.id;
  return NULL;
end;
$$;



CREATE FUNCTION del_recurring() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
  delete from recurring where id = old.id;
  delete from recurringemail where id = old.id;
  delete from recurringprint where id = old.id;
  return NULL;
end;
$$;



CREATE FUNCTION del_vendor() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
  delete from shipto where trans_id = old.id;
  delete from vendortax where vendor_id = old.id;
  delete from partsvendor where vendor_id = old.id;
  delete from address where trans_id = old.id;
  return NULL;
end;
$$;



CREATE FUNCTION del_yearend() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
  delete from yearend where trans_id = old.id;
  return NULL;
end;
$$;



CREATE FUNCTION lastcost(integer) RETURNS double precision
    LANGUAGE plpgsql
    AS $_$
 
DECLARE 
  
v_cost float; 
v_parts_id alias for $1; 
   
BEGIN 
    
  SELECT INTO v_cost sellprice FROM invoice i
  JOIN ap a ON (a.id = i.trans_id) 
  WHERE i.parts_id = v_parts_id 
  ORDER BY a.transdate desc, a.id desc 
  LIMIT 1; 
 
  IF v_cost IS NULL THEN 
    v_cost := 0; 
  END IF; 

RETURN v_cost; 
END; 
$_$;



CREATE FUNCTION to_filtered_tsvector(input text, filter_type text DEFAULT 'COMMON'::text, ts_config regconfig DEFAULT 'simple'::regconfig) RETURNS tsvector
    LANGUAGE plpgsql
    AS $$ DECLARE filtered_input TEXT; BEGIN SELECT string_agg(input_word, ' ') INTO filtered_input FROM unnest(string_to_array(input, ' ')) AS input_word WHERE lower(input_word) NOT IN (SELECT lower(word) FROM search_irrelevant_words WHERE search_target = 'COMMON' OR search_target = filter_type); RETURN to_tsvector(ts_config, filtered_input); END; $$;



CREATE FUNCTION update_entity_paths() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Self-reference
    INSERT INTO hierarchy_paths (ancestor_id, descendant_id, depth)
    VALUES (NEW.id, NEW.id, 0);
    
    -- Parent paths
    IF NEW.parent_id IS NOT NULL THEN
        INSERT INTO hierarchy_paths (ancestor_id, descendant_id, depth)
        SELECT p.ancestor_id, NEW.id, p.depth + 1
        FROM hierarchy_paths p
        WHERE p.descendant_id = NEW.parent_id;
    END IF;
    
    RETURN NEW;
END;
$$;



CREATE FUNCTION update_timestamps() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;



CREATE SEQUENCE entry_id
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


SET default_tablespace = '';

SET default_table_access_method = heap;


CREATE TABLE acc_trans (
    trans_id integer,
    chart_id integer NOT NULL,
    amount double precision,
    transdate date DEFAULT CURRENT_DATE,
    source text,
    approved boolean DEFAULT true,
    fx_transaction boolean DEFAULT false,
    project_id integer,
    memo text,
    id integer,
    cleared boolean DEFAULT false,
    vr_id integer,
    entry_id integer DEFAULT nextval('entry_id'::regclass),
    tax text,
    taxamount double precision,
    tax_chart_id integer,
    invoice_id integer,
    reconciled date
);



CREATE TABLE acc_trans_log (
    trans_id integer,
    chart_id integer,
    amount double precision,
    transdate date,
    source text,
    approved boolean,
    fx_transaction boolean,
    project_id integer,
    memo text,
    id integer,
    cleared boolean,
    vr_id integer,
    entry_id integer,
    tax text,
    taxamount double precision,
    tax_chart_id integer,
    invoice_id integer,
    reconciled date,
    ts timestamp without time zone DEFAULT now()
);



CREATE TABLE acc_trans_log_deleted (
    trans_id integer,
    chart_id integer,
    amount double precision,
    transdate date,
    source text,
    approved boolean,
    fx_transaction boolean,
    project_id integer,
    memo text,
    id integer,
    cleared boolean,
    vr_id integer,
    entry_id integer,
    tax text,
    taxamount double precision,
    tax_chart_id integer,
    invoice_id integer,
    reconciled date,
    ts timestamp without time zone DEFAULT now()
);



CREATE SEQUENCE addressid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



CREATE TABLE address (
    id integer DEFAULT nextval('addressid'::regclass) NOT NULL,
    trans_id integer,
    address1 character varying(64),
    address2 character varying(64),
    city character varying(64),
    is_migrated boolean DEFAULT true,
    post_office character varying(64),
    state character varying(32),
    zipcode character varying(32),
    country character varying(32)
);



CREATE SEQUENCE id
    START WITH 10000
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



CREATE TABLE ap (
    id integer DEFAULT nextval('id'::regclass),
    invnumber text,
    transdate date DEFAULT CURRENT_DATE,
    vendor_id integer,
    taxincluded boolean DEFAULT false,
    amount double precision,
    netamount double precision,
    paid double precision,
    datepaid date,
    duedate date,
    invoice boolean DEFAULT false,
    ordnumber text,
    curr character(3),
    notes text,
    till character varying(20),
    quonumber text,
    intnotes text,
    department_id integer DEFAULT 0,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    shippingpoint text,
    terms smallint DEFAULT 0,
    approved boolean DEFAULT true,
    cashdiscount real,
    discountterms smallint,
    waybill text,
    warehouse_id integer,
    linetax boolean DEFAULT false,
    ts timestamp without time zone DEFAULT now(),
    exchangerate double precision,
    dcn text
);



CREATE TABLE ap_log (
    id integer,
    invnumber text,
    transdate date,
    vendor_id integer,
    taxincluded boolean,
    amount double precision,
    netamount double precision,
    paid double precision,
    datepaid date,
    duedate date,
    invoice boolean,
    ordnumber text,
    curr character(3),
    notes text,
    till character varying(20),
    quonumber text,
    intnotes text,
    department_id integer,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    shippingpoint text,
    terms smallint,
    approved boolean,
    cashdiscount real,
    discountterms smallint,
    waybill text,
    warehouse_id integer,
    linetax boolean,
    ts timestamp without time zone
);



CREATE TABLE ap_log_deleted (
    id integer,
    invnumber text,
    transdate date,
    vendor_id integer,
    taxincluded boolean,
    amount double precision,
    netamount double precision,
    paid double precision,
    datepaid date,
    duedate date,
    invoice boolean,
    ordnumber text,
    curr character(3),
    notes text,
    till character varying(20),
    quonumber text,
    intnotes text,
    department_id integer,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    shippingpoint text,
    terms smallint,
    approved boolean,
    cashdiscount real,
    discountterms smallint,
    waybill text,
    warehouse_id integer,
    linetax boolean,
    ts timestamp without time zone
);



CREATE TABLE ar (
    id integer DEFAULT nextval('id'::regclass),
    invnumber text,
    transdate date DEFAULT CURRENT_DATE,
    customer_id integer,
    taxincluded boolean DEFAULT false,
    amount double precision,
    netamount double precision,
    paid double precision,
    datepaid date,
    duedate date,
    invoice boolean DEFAULT false,
    shippingpoint text,
    terms smallint,
    notes text,
    curr character(3),
    ordnumber text,
    till character varying(20),
    quonumber text,
    intnotes text,
    department_id integer DEFAULT 0,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    approved boolean DEFAULT true,
    cashdiscount real,
    discountterms smallint,
    waybill text,
    warehouse_id integer,
    linetax boolean DEFAULT false,
    ts timestamp without time zone DEFAULT now(),
    exchangerate double precision,
    dcn text
);



CREATE TABLE ar_log (
    id integer,
    invnumber text,
    transdate date,
    customer_id integer,
    taxincluded boolean,
    amount double precision,
    netamount double precision,
    paid double precision,
    datepaid date,
    duedate date,
    invoice boolean,
    shippingpoint text,
    terms smallint,
    notes text,
    curr character(3),
    ordnumber text,
    till character varying(20),
    quonumber text,
    intnotes text,
    department_id integer,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    approved boolean,
    cashdiscount real,
    discountterms smallint,
    waybill text,
    warehouse_id integer,
    linetax boolean,
    ts timestamp without time zone
);



CREATE TABLE ar_log_deleted (
    id integer,
    invnumber text,
    transdate date,
    customer_id integer,
    taxincluded boolean,
    amount double precision,
    netamount double precision,
    paid double precision,
    datepaid date,
    duedate date,
    invoice boolean,
    shippingpoint text,
    terms smallint,
    notes text,
    curr character(3),
    ordnumber text,
    till character varying(20),
    quonumber text,
    intnotes text,
    department_id integer,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    approved boolean,
    cashdiscount real,
    discountterms smallint,
    waybill text,
    warehouse_id integer,
    linetax boolean,
    ts timestamp without time zone
);



CREATE SEQUENCE assemblyid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



CREATE TABLE assembly (
    id integer DEFAULT nextval('assemblyid'::regclass),
    parts_id integer,
    qty double precision,
    bom boolean,
    adj boolean,
    aid integer
);



CREATE TABLE audittrail (
    trans_id integer,
    tablename text,
    reference text,
    formname text,
    action text,
    transdate timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    employee_id integer
);



CREATE TABLE bank (
    id integer,
    name character varying(64),
    iban character varying(34),
    bic character varying(11),
    address_id integer DEFAULT nextval('addressid'::regclass),
    dcn text,
    rvc text,
    strdbkginf text,
    invdescriptionqr text,
    qriban text,
    membernumber text
);



CREATE TABLE bank_account (
    id integer NOT NULL,
    bic text,
    iban text,
    account text,
    bank text,
    branch text,
    country text,
    country_iso text,
    city text,
    state text,
    zip text,
    address text,
    description text,
    chart_id integer,
    inactive boolean,
    blink_permission_id bigint,
    blink_account_id text
);



ALTER TABLE bank_account ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME bank_account_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);



CREATE TABLE banking_import_event (
    id bigint NOT NULL,
    delegated_to text
);



ALTER TABLE banking_import_event ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME banking_import_event_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);



CREATE TABLE blink_import_process (
    id bigint NOT NULL,
    bank_account_id bigint NOT NULL,
    status text NOT NULL,
    error jsonb,
    created_at timestamp without time zone NOT NULL,
    last_modified_at timestamp without time zone
);



ALTER TABLE blink_import_process ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME blink_import_process_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);



CREATE TABLE blink_import_process_log (
    id bigint NOT NULL,
    blink_import_process_id bigint NOT NULL,
    processed_target_id text NOT NULL,
    processed_payload jsonb NOT NULL,
    error jsonb,
    banking_import_event_id bigint
);



ALTER TABLE blink_import_process_log ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME blink_import_process_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);



CREATE TABLE booking_to_settlement (
    id integer NOT NULL,
    booking_id integer NOT NULL,
    settlement_id integer NOT NULL
);



CREATE SEQUENCE booking_to_settlement_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE booking_to_settlement_id_seq OWNED BY booking_to_settlement.id;



CREATE TABLE br (
    id integer DEFAULT nextval('id'::regclass) NOT NULL,
    batchnumber text,
    description text,
    batch text,
    transdate date DEFAULT CURRENT_DATE,
    apprdate date,
    amount double precision,
    managerid integer,
    employee_id integer
);



CREATE TABLE build (
    id integer DEFAULT nextval('id'::regclass) NOT NULL,
    reference text,
    transdate date,
    department_id integer,
    warehouse_id integer,
    employee_id integer
);



CREATE TABLE business (
    id integer DEFAULT nextval('id'::regclass),
    description text,
    discount real
);



CREATE TABLE cargo (
    id integer NOT NULL,
    trans_id integer NOT NULL,
    package text,
    netweight double precision,
    grossweight double precision,
    volume double precision
);



CREATE TABLE chart (
    id integer DEFAULT nextval('id'::regclass),
    accno integer,
    description text,
    balance double precision,
    type character(1),
    gifi integer,
    category character(1),
    link text,
    gifi_accno text,
    contra boolean DEFAULT false,
    parent_id integer,
    charttype character(1) DEFAULT 'A'::bpchar
);



CREATE TABLE chat (
    id integer NOT NULL,
    trans_id integer NOT NULL,
    message character varying(255) NOT NULL,
    employee_id integer NOT NULL,
    creation_date timestamp without time zone NOT NULL
);



CREATE SEQUENCE chat_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE chat_id_seq OWNED BY chat.id;



CREATE SEQUENCE contactid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



CREATE TABLE contact (
    id integer DEFAULT nextval('contactid'::regclass) NOT NULL,
    trans_id integer NOT NULL,
    salutation character varying(32),
    firstname character varying(32),
    lastname character varying(32),
    contacttitle character varying(32),
    occupation character varying(32),
    phone character varying(20),
    fax character varying(20),
    mobile character varying(20),
    email text,
    gender character(1) DEFAULT 'M'::bpchar,
    parent_id integer,
    typeofcontact character varying(20)
);



CREATE TABLE credits (
    id integer NOT NULL,
    reference text,
    description text,
    transdate date,
    accno text,
    amount numeric(12,2)
);



CREATE SEQUENCE credits_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE credits_id_seq OWNED BY credits.id;



CREATE TABLE curr (
    rn integer,
    curr character(3) NOT NULL,
    "precision" smallint
);



CREATE TABLE customer (
    id integer DEFAULT nextval('id'::regclass) NOT NULL,
    name character varying(64),
    contact character varying(64),
    phone character varying(20),
    fax character varying(20),
    email text,
    notes text,
    terms smallint DEFAULT 0,
    taxincluded boolean DEFAULT false,
    customernumber character varying(32),
    cc text,
    bcc text,
    business_id integer,
    taxnumber character varying(32),
    sic_code character varying(6),
    discount real,
    creditlimit double precision DEFAULT 0,
    iban character varying(34),
    bic character varying(11),
    employee_id integer,
    language_code character varying(6),
    pricegroup_id integer,
    curr character(3),
    startdate date,
    enddate date,
    arap_accno_id integer,
    payment_accno_id integer,
    discount_accno_id integer,
    cashdiscount real,
    discountterms smallint,
    threshold double precision,
    dispatch_id integer
);



CREATE TABLE customercart (
    cart_id character varying(32),
    customer_id integer,
    parts_id integer,
    qty double precision DEFAULT 0,
    price double precision DEFAULT 0,
    taxaccounts text
);



CREATE SEQUENCE customerloginid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



CREATE TABLE customerlogin (
    id integer DEFAULT nextval('customerloginid'::regclass) NOT NULL,
    login character varying(100),
    passwd character(32),
    session character(32),
    session_exp character varying(20),
    customer_id integer
);



CREATE TABLE customertax (
    customer_id integer,
    chart_id integer
);



CREATE TABLE debits (
    id integer NOT NULL,
    reference text,
    description text,
    transdate date,
    accno text,
    amount numeric(12,2)
);



CREATE SEQUENCE debits_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE debits_id_seq OWNED BY debits.id;



CREATE TABLE debitscredits (
    id integer NOT NULL,
    reference text,
    description text,
    transdate date,
    debit_accno text,
    credit_accno text,
    amount numeric(12,2)
);



CREATE SEQUENCE debitscredits_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE debitscredits_id_seq OWNED BY debitscredits.id;



CREATE TABLE defaults (
    fldname text,
    fldvalue text
);



CREATE TABLE department (
    id integer DEFAULT nextval('id'::regclass),
    description text,
    role character(1) DEFAULT 'P'::bpchar
);



CREATE TABLE dispatch (
    id integer DEFAULT nextval('id'::regclass),
    description text
);



CREATE TABLE dpt_trans (
    trans_id integer,
    department_id integer
);



CREATE TABLE employee (
    id integer DEFAULT nextval('id'::regclass),
    login text,
    name character varying(64),
    address1 character varying(32),
    address2 character varying(32),
    city character varying(32),
    state character varying(32),
    zipcode character varying(10),
    country character varying(32),
    workphone character varying(20),
    homephone character varying(20),
    startdate date DEFAULT CURRENT_DATE,
    enddate date,
    notes text,
    role character varying(20),
    sales boolean DEFAULT false,
    email text,
    ssn character varying(20),
    iban character varying(34),
    bic character varying(11),
    managerid integer,
    employeenumber character varying(32),
    dob date
);



CREATE TABLE entities (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    type entity_type NOT NULL,
    parent_id integer,
    organization_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);



CREATE SEQUENCE entities_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE entities_id_seq OWNED BY entities.id;



CREATE TABLE exchangerate (
    curr character(3),
    transdate date,
    buy double precision,
    sell double precision
);



CREATE TABLE fifo (
    trans_id integer,
    transdate date,
    parts_id integer,
    qty double precision,
    costprice double precision,
    sellprice double precision,
    warehouse_id integer,
    invoice_id integer
);



CREATE TABLE filtered (
    trans_id integer,
    entry_id integer DEFAULT 0,
    entry_id2 integer DEFAULT 0,
    debit double precision DEFAULT 0,
    credit double precision DEFAULT 0
);



CREATE TABLE financial_year (
    id integer NOT NULL,
    start_date date,
    end_date date,
    status financial_year_status_enum,
    yearend_id integer
);



CREATE SEQUENCE financial_year_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE financial_year_id_seq OWNED BY financial_year.id;



CREATE TABLE gifi (
    accno text,
    description text
);



CREATE TABLE gl (
    id integer DEFAULT nextval('id'::regclass),
    reference text,
    description text,
    transdate date DEFAULT CURRENT_DATE,
    employee_id integer,
    notes text,
    department_id integer DEFAULT 0,
    approved boolean DEFAULT true,
    curr character(3),
    exchangerate double precision,
    ts timestamp without time zone DEFAULT now(),
    onhold boolean
);



CREATE TABLE gl_log (
    id integer,
    reference text,
    description text,
    transdate date,
    employee_id integer,
    notes text,
    department_id integer,
    approved boolean,
    curr character(3),
    exchangerate double precision,
    ts timestamp without time zone,
    onhold boolean
);



CREATE TABLE gl_log_deleted (
    id integer,
    reference text,
    description text,
    transdate date,
    employee_id integer,
    notes text,
    department_id integer,
    approved boolean,
    curr character(3),
    exchangerate double precision,
    ts timestamp without time zone,
    onhold boolean
);



CREATE TABLE hierarchy_paths (
    ancestor_id integer NOT NULL,
    descendant_id integer NOT NULL,
    depth integer NOT NULL
);



CREATE SEQUENCE inventoryid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



CREATE TABLE inventory (
    id integer DEFAULT nextval('inventoryid'::regclass),
    warehouse_id integer,
    parts_id integer,
    trans_id integer,
    orderitems_id integer,
    qty double precision,
    shippingdate date,
    employee_id integer,
    department_id integer,
    warehouse_id2 integer,
    serialnumber text,
    itemnotes text,
    cost double precision,
    linetype character(1) DEFAULT '0'::bpchar,
    description text,
    invoice_id integer,
    cogs double precision
);



CREATE SEQUENCE invoiceid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



CREATE TABLE invoice (
    id integer DEFAULT nextval('invoiceid'::regclass),
    trans_id integer,
    parts_id integer,
    description text,
    qty real,
    allocated real,
    sellprice double precision,
    fxsellprice double precision,
    discount real,
    assemblyitem boolean DEFAULT false,
    project_id integer,
    deliverydate date,
    serialnumber text,
    notes text
);



CREATE TABLE invoice_log (
    id integer,
    trans_id integer,
    parts_id integer,
    description text,
    qty real,
    allocated real,
    sellprice double precision,
    fxsellprice double precision,
    discount real,
    assemblyitem boolean,
    project_id integer,
    deliverydate date,
    serialnumber text,
    notes text,
    ts timestamp without time zone DEFAULT now()
);



CREATE TABLE invoice_log_deleted (
    trans_id integer,
    chart_id integer,
    amount double precision,
    transdate date,
    source text,
    approved boolean,
    fx_transaction boolean,
    project_id integer,
    memo text,
    id integer,
    cleared boolean,
    vr_id integer,
    entry_id integer,
    tax text,
    taxamount double precision,
    tax_chart_id integer,
    invoice_id integer,
    reconciled date,
    ts timestamp without time zone DEFAULT now()
);



CREATE TABLE invoicetax (
    trans_id integer,
    invoice_id integer,
    chart_id integer,
    taxamount double precision,
    amount double precision
);



CREATE SEQUENCE jcitemsid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



CREATE TABLE jcitems (
    id integer DEFAULT nextval('jcitemsid'::regclass),
    project_id integer,
    parts_id integer,
    description text,
    qty double precision,
    allocated double precision,
    sellprice double precision,
    fxsellprice double precision,
    serialnumber text,
    checkedin timestamp with time zone,
    checkedout timestamp with time zone,
    employee_id integer,
    notes text
);



CREATE TABLE language (
    code character varying(6),
    description text
);



CREATE TABLE lastused (
    id integer NOT NULL,
    report character varying(40),
    cols text,
    login character varying(255)
);



CREATE SEQUENCE lastused_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE lastused_id_seq OWNED BY lastused.id;



CREATE TABLE makemodel (
    parts_id integer,
    make text,
    model text
);



CREATE TABLE oe (
    id integer DEFAULT nextval('id'::regclass),
    ordnumber text,
    transdate date DEFAULT CURRENT_DATE,
    vendor_id integer,
    customer_id integer,
    amount double precision,
    netamount double precision,
    reqdate date,
    taxincluded boolean,
    shippingpoint text,
    notes text,
    curr character(3),
    employee_id integer,
    closed boolean DEFAULT false,
    quotation boolean DEFAULT false,
    quonumber text,
    intnotes text,
    department_id integer DEFAULT 0,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    terms smallint DEFAULT 0,
    waybill text,
    warehouse_id integer,
    description text,
    aa_id integer,
    exchangerate double precision
);



CREATE TABLE oldchart (
    id integer DEFAULT nextval('id'::regclass),
    accno text NOT NULL,
    description text,
    parent_id integer,
    charttype character(1) DEFAULT 'A'::bpchar,
    category character(1),
    link text,
    gifi_accno text,
    contra boolean DEFAULT false,
    allow_gl boolean DEFAULT true,
    symbol_link character varying(128)
);



CREATE SEQUENCE orderitemsid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



CREATE TABLE orderitems (
    id integer DEFAULT nextval('orderitemsid'::regclass),
    trans_id integer,
    parts_id integer,
    description text,
    qty double precision,
    sellprice double precision,
    discount real,
    unit character varying(5),
    project_id integer,
    reqdate date,
    ship double precision,
    serialnumber text,
    itemnotes text,
    lineitemdetail boolean,
    ordernumber text,
    ponumber text,
    notes text
);



CREATE TABLE org (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    hierarchy_level integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);



CREATE SEQUENCE org_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE org_id_seq OWNED BY org.id;



CREATE TABLE parts (
    id integer DEFAULT nextval('id'::regclass),
    partnumber text,
    description text,
    unit character varying(5),
    listprice double precision,
    sellprice double precision,
    lastcost double precision,
    priceupdate date DEFAULT CURRENT_DATE,
    weight double precision,
    onhand double precision DEFAULT 0,
    notes text,
    makemodel boolean DEFAULT false,
    assembly boolean DEFAULT false,
    alternate boolean DEFAULT false,
    rop double precision,
    inventory_accno_id integer,
    income_accno_id integer,
    expense_accno_id integer,
    bin text,
    obsolete boolean DEFAULT false,
    bom boolean DEFAULT false,
    image text,
    drawing text,
    microfiche text,
    partsgroup_id integer,
    project_id integer,
    avgcost double precision,
    tariff_hscode text,
    countryorigin text,
    barcode text,
    toolnumber text
);



CREATE TABLE partsattr (
    parts_id integer,
    hotnew character varying(3)
);



CREATE TABLE partscustomer (
    parts_id integer,
    customer_id integer,
    pricegroup_id integer,
    pricebreak double precision,
    sellprice double precision,
    validfrom date,
    validto date,
    curr character(3)
);



CREATE TABLE partsgroup (
    id integer DEFAULT nextval('id'::regclass),
    partsgroup text,
    pos boolean DEFAULT true
);



CREATE TABLE partstax (
    parts_id integer,
    chart_id integer
);



CREATE TABLE partsvendor (
    vendor_id integer,
    parts_id integer,
    partnumber text,
    leadtime smallint,
    lastcost double precision,
    curr character(3)
);



CREATE TABLE payment (
    id integer NOT NULL,
    trans_id integer NOT NULL,
    exchangerate double precision DEFAULT 1,
    paymentmethod_id integer
);



CREATE TABLE paymentmethod (
    id integer DEFAULT nextval('id'::regclass) NOT NULL,
    description text,
    fee double precision,
    rn integer
);



CREATE TABLE pricegroup (
    id integer DEFAULT nextval('id'::regclass),
    pricegroup text
);



CREATE TABLE project (
    id integer DEFAULT nextval('id'::regclass),
    projectnumber text,
    description text,
    startdate date,
    enddate date,
    parts_id integer,
    production double precision DEFAULT 0,
    completed double precision DEFAULT 0,
    customer_id integer
);



CREATE TABLE recurring (
    id integer,
    reference text,
    startdate date,
    nextdate date,
    enddate date,
    repeat smallint,
    unit character varying(6),
    howmany integer,
    payment boolean DEFAULT false,
    description text
);



CREATE TABLE recurringemail (
    id integer,
    formname text,
    format text,
    message text
);



CREATE TABLE recurringprint (
    id integer,
    formname text,
    format text,
    printer text
);



CREATE TABLE report (
    reportid integer DEFAULT nextval('id'::regclass) NOT NULL,
    reportcode text,
    reportdescription text,
    login text
);



CREATE TABLE reportvars (
    reportid integer NOT NULL,
    reportvariable text,
    reportvalue text
);



CREATE TABLE schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);



CREATE TABLE search_irrelevant_words (
    id integer NOT NULL,
    search_target text NOT NULL,
    word text NOT NULL
);



CREATE SEQUENCE search_irrelevant_words_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE search_irrelevant_words_id_seq OWNED BY search_irrelevant_words.id;



CREATE TABLE semaphore (
    id integer,
    login text,
    module text,
    expires character varying(10)
);



CREATE TABLE shipto (
    trans_id integer,
    shiptoname character varying(64),
    shiptoaddress1 character varying(32),
    shiptoaddress2 character varying(32),
    shiptocity character varying(32),
    shiptostate character varying(32),
    shiptozipcode character varying(10),
    shiptocountry character varying(32),
    shiptocontact character varying(64),
    shiptophone character varying(20),
    shiptofax character varying(20),
    shiptoemail text
);



CREATE TABLE sic (
    code character varying(6),
    sictype character(1),
    description text
);



CREATE TABLE status (
    trans_id integer,
    formname text,
    printed boolean DEFAULT false,
    emailed boolean DEFAULT false,
    spoolfile text
);



CREATE TABLE tax (
    chart_id integer,
    rate double precision,
    taxnumber text,
    reversecharge_id integer,
    validto date,
    vatkey character varying(5),
    formdigit integer DEFAULT 0,
    validfrom date,
    id integer NOT NULL
);



CREATE SEQUENCE tax_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE tax_id_seq OWNED BY tax.id;



CREATE TABLE translation (
    trans_id integer,
    language_code character varying(6),
    description text
);



CREATE TABLE trf (
    id integer DEFAULT nextval('id'::regclass) NOT NULL,
    transdate date,
    trfnumber text,
    description text,
    notes text,
    department_id integer,
    from_warehouse_id integer,
    to_warehouse_id integer DEFAULT 0,
    employee_id integer DEFAULT 0,
    delivereddate date
);



CREATE TABLE users (
    id integer NOT NULL,
    email character varying(255) NOT NULL,
    full_name character varying(255) NOT NULL,
    entity_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    last_login_at timestamp with time zone,
    deleted_at timestamp with time zone
);



CREATE SEQUENCE users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE users_id_seq OWNED BY users.id;



CREATE VIEW v_active_users AS
 SELECT u.id AS user_id,
    u.email,
    u.full_name,
    o.name AS organization,
    c.name AS company,
    r.name AS regional,
    d.name AS department,
    cc.name AS cost_center,
    u.last_login_at
   FROM (((((users u
     JOIN entities cc ON ((u.entity_id = cc.id)))
     JOIN entities d ON ((cc.parent_id = d.id)))
     JOIN entities r ON ((d.parent_id = r.id)))
     JOIN entities c ON ((r.parent_id = c.id)))
     JOIN org o ON ((c.organization_id = o.id)))
  WHERE ((u.deleted_at IS NULL) AND (u.last_login_at > (CURRENT_DATE - '90 days'::interval)) AND (cc.deleted_at IS NULL) AND (d.deleted_at IS NULL) AND (r.deleted_at IS NULL) AND (c.deleted_at IS NULL) AND (o.deleted_at IS NULL));



CREATE VIEW v_company_subtree AS
 SELECT c.id AS company_id,
    c.name AS company,
    e.id AS entity_id,
    e.name AS entity_name,
    e.type AS entity_type,
    hp.depth AS levels_from_company
   FROM ((entities c
     JOIN hierarchy_paths hp ON ((c.id = hp.ancestor_id)))
     JOIN entities e ON ((hp.descendant_id = e.id)))
  WHERE ((c.type = 'COMPANY'::entity_type) AND (c.deleted_at IS NULL) AND (e.deleted_at IS NULL));



CREATE VIEW v_cost_center_management AS
 SELECT cc.id AS cost_center_id,
    cc.name AS cost_center,
    d.id AS department_id,
    d.name AS department,
    r.id AS regional_id,
    r.name AS regional,
    c.id AS company_id,
    c.name AS company,
    o.id AS organization_id,
    o.name AS organization,
    count(u.id) AS user_count,
    string_agg((u.full_name)::text, ', '::text) AS users
   FROM (((((entities cc
     JOIN entities d ON ((cc.parent_id = d.id)))
     JOIN entities r ON ((d.parent_id = r.id)))
     JOIN entities c ON ((r.parent_id = c.id)))
     JOIN org o ON ((c.organization_id = o.id)))
     LEFT JOIN users u ON ((u.entity_id = cc.id)))
  WHERE ((cc.type = 'COST_CENTER'::entity_type) AND (cc.deleted_at IS NULL))
  GROUP BY cc.id, cc.name, d.id, d.name, r.id, r.name, c.id, c.name, o.id, o.name;



CREATE VIEW v_department_summary AS
 SELECT d.id AS department_id,
    d.name AS department,
    r.id AS regional_id,
    r.name AS regional,
    c.id AS company_id,
    c.name AS company,
    o.id AS organization_id,
    o.name AS organization,
    count(DISTINCT cc.id) AS cost_center_count,
    count(DISTINCT u.id) AS user_count
   FROM (((((entities d
     JOIN entities r ON ((d.parent_id = r.id)))
     JOIN entities c ON ((r.parent_id = c.id)))
     JOIN org o ON ((c.organization_id = o.id)))
     LEFT JOIN entities cc ON (((cc.parent_id = d.id) AND (cc.type = 'COST_CENTER'::entity_type) AND (cc.deleted_at IS NULL))))
     LEFT JOIN users u ON (((u.entity_id = cc.id) AND (u.deleted_at IS NULL))))
  WHERE ((d.type = 'DEPARTMENT'::entity_type) AND (d.deleted_at IS NULL))
  GROUP BY d.id, d.name, r.id, r.name, c.id, c.name, o.id, o.name;



CREATE VIEW v_organization_hierarchy AS
 WITH RECURSIVE org_chart AS (
         SELECT entities.id,
            entities.name,
            entities.type,
            entities.parent_id,
            entities.organization_id,
            entities.deleted_at,
            (entities.name)::text AS path,
            0 AS depth
           FROM entities
          WHERE (entities.parent_id IS NULL)
        UNION ALL
         SELECT e.id,
            e.name,
            e.type,
            e.parent_id,
            e.organization_id,
            e.deleted_at,
            ((oc_1.path || ' > '::text) || (e.name)::text) AS text,
            (oc_1.depth + 1)
           FROM (entities e
             JOIN org_chart oc_1 ON ((e.parent_id = oc_1.id)))
        )
 SELECT o.name AS organization,
    oc.id AS entity_id,
    oc.name AS entity_name,
    oc.type AS entity_type,
    oc.path AS full_path,
    oc.depth
   FROM (org_chart oc
     JOIN org o ON ((oc.organization_id = o.id)))
  WHERE (oc.deleted_at IS NULL);



CREATE VIEW v_organization_size_report AS
 SELECT o.id AS organization_id,
    o.name AS organization,
    count(DISTINCT c.id) FILTER (WHERE ((c.type = 'COMPANY'::entity_type) AND (c.deleted_at IS NULL))) AS company_count,
    count(DISTINCT r.id) FILTER (WHERE ((r.type = 'REGIONAL'::entity_type) AND (r.deleted_at IS NULL))) AS regional_count,
    count(DISTINCT d.id) FILTER (WHERE ((d.type = 'DEPARTMENT'::entity_type) AND (d.deleted_at IS NULL))) AS department_count,
    count(DISTINCT cc.id) FILTER (WHERE ((cc.type = 'COST_CENTER'::entity_type) AND (cc.deleted_at IS NULL))) AS cost_center_count,
    count(DISTINCT u.id) FILTER (WHERE (u.deleted_at IS NULL)) AS user_count
   FROM (((((org o
     LEFT JOIN entities c ON ((c.organization_id = o.id)))
     LEFT JOIN entities r ON ((r.organization_id = o.id)))
     LEFT JOIN entities d ON ((d.organization_id = o.id)))
     LEFT JOIN entities cc ON ((cc.organization_id = o.id)))
     LEFT JOIN users u ON ((u.entity_id = cc.id)))
  WHERE (o.deleted_at IS NULL)
  GROUP BY o.id, o.name;



CREATE VIEW v_user_directory AS
 SELECT u.id AS user_id,
    u.email,
    u.full_name,
    u.last_login_at,
    cc.id AS cost_center_id,
    cc.name AS cost_center,
    d.id AS department_id,
    d.name AS department,
    r.id AS regional_id,
    r.name AS regional,
    c.id AS company_id,
    c.name AS company,
    o.id AS organization_id,
    o.name AS organization
   FROM (((((users u
     JOIN entities cc ON ((u.entity_id = cc.id)))
     JOIN entities d ON ((cc.parent_id = d.id)))
     JOIN entities r ON ((d.parent_id = r.id)))
     JOIN entities c ON ((r.parent_id = c.id)))
     JOIN org o ON ((c.organization_id = o.id)))
  WHERE (u.deleted_at IS NULL);



CREATE TABLE vat_settlement (
    id integer NOT NULL,
    vat_form_id integer,
    period_from date,
    period_to date,
    creation_date timestamp without time zone,
    data json,
    xml bytea
);



CREATE SEQUENCE vat_settlement_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE vat_settlement_id_seq OWNED BY vat_settlement.id;



CREATE TABLE vendor (
    id integer DEFAULT nextval('id'::regclass) NOT NULL,
    name character varying(64),
    contact character varying(64),
    phone character varying(20),
    fax character varying(20),
    email text,
    notes text,
    terms smallint DEFAULT 0,
    taxincluded boolean DEFAULT false,
    vendornumber character varying(32),
    cc text,
    bcc text,
    gifi_accno character varying(30),
    business_id integer,
    taxnumber character varying(32),
    sic_code character varying(6),
    discount real,
    creditlimit double precision DEFAULT 0,
    iban character varying(34),
    bic character varying(11),
    employee_id integer,
    language_code character varying(6),
    pricegroup_id integer,
    curr character(3),
    startdate date,
    enddate date,
    arap_accno_id integer,
    payment_accno_id integer,
    discount_accno_id integer,
    cashdiscount real,
    discountterms smallint,
    threshold double precision,
    dispatch_id integer
);



CREATE TABLE vendortax (
    vendor_id integer,
    chart_id integer
);



CREATE TABLE vr (
    br_id integer,
    trans_id integer NOT NULL,
    id integer DEFAULT nextval('id'::regclass) NOT NULL,
    vouchernumber text
);



CREATE TABLE warehouse (
    id integer DEFAULT nextval('id'::regclass),
    description text
);



CREATE TABLE yearend (
    trans_id integer,
    transdate date,
    id integer NOT NULL
);



CREATE SEQUENCE yearend_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;



ALTER SEQUENCE yearend_id_seq OWNED BY yearend.id;



ALTER TABLE ONLY booking_to_settlement ALTER COLUMN id SET DEFAULT nextval('booking_to_settlement_id_seq'::regclass);



ALTER TABLE ONLY chat ALTER COLUMN id SET DEFAULT nextval('chat_id_seq'::regclass);



ALTER TABLE ONLY credits ALTER COLUMN id SET DEFAULT nextval('credits_id_seq'::regclass);



ALTER TABLE ONLY debits ALTER COLUMN id SET DEFAULT nextval('debits_id_seq'::regclass);



ALTER TABLE ONLY debitscredits ALTER COLUMN id SET DEFAULT nextval('debitscredits_id_seq'::regclass);



ALTER TABLE ONLY entities ALTER COLUMN id SET DEFAULT nextval('entities_id_seq'::regclass);



ALTER TABLE ONLY financial_year ALTER COLUMN id SET DEFAULT nextval('financial_year_id_seq'::regclass);



ALTER TABLE ONLY lastused ALTER COLUMN id SET DEFAULT nextval('lastused_id_seq'::regclass);



ALTER TABLE ONLY org ALTER COLUMN id SET DEFAULT nextval('org_id_seq'::regclass);



ALTER TABLE ONLY search_irrelevant_words ALTER COLUMN id SET DEFAULT nextval('search_irrelevant_words_id_seq'::regclass);



ALTER TABLE ONLY tax ALTER COLUMN id SET DEFAULT nextval('tax_id_seq'::regclass);



ALTER TABLE ONLY users ALTER COLUMN id SET DEFAULT nextval('users_id_seq'::regclass);



ALTER TABLE ONLY vat_settlement ALTER COLUMN id SET DEFAULT nextval('vat_settlement_id_seq'::regclass);



ALTER TABLE ONLY yearend ALTER COLUMN id SET DEFAULT nextval('yearend_id_seq'::regclass);



ALTER TABLE ONLY address
    ADD CONSTRAINT address_pkey PRIMARY KEY (id);



ALTER TABLE ONLY bank_account
    ADD CONSTRAINT bank_account_pkey PRIMARY KEY (id);



ALTER TABLE ONLY banking_import_event
    ADD CONSTRAINT banking_import_event_pkey PRIMARY KEY (id);



ALTER TABLE ONLY blink_import_process_log
    ADD CONSTRAINT blink_import_process_log_pkey PRIMARY KEY (id);



ALTER TABLE ONLY blink_import_process
    ADD CONSTRAINT blink_import_process_pkey PRIMARY KEY (id);



ALTER TABLE ONLY booking_to_settlement
    ADD CONSTRAINT booking_to_settlement_pkey PRIMARY KEY (id);



ALTER TABLE ONLY br
    ADD CONSTRAINT br_pkey PRIMARY KEY (id);



ALTER TABLE ONLY build
    ADD CONSTRAINT build_pkey PRIMARY KEY (id);



ALTER TABLE ONLY chart
    ADD CONSTRAINT chart_accno_key1 UNIQUE (accno);



ALTER TABLE ONLY chat
    ADD CONSTRAINT chat_pkey PRIMARY KEY (id);



ALTER TABLE ONLY contact
    ADD CONSTRAINT contact_pkey PRIMARY KEY (id);



ALTER TABLE ONLY curr
    ADD CONSTRAINT curr_pkey PRIMARY KEY (curr);



ALTER TABLE ONLY customer
    ADD CONSTRAINT customer_pkey PRIMARY KEY (id);



ALTER TABLE ONLY customerlogin
    ADD CONSTRAINT customerlogin_login_key UNIQUE (login);



ALTER TABLE ONLY customerlogin
    ADD CONSTRAINT customerlogin_pkey PRIMARY KEY (id);



ALTER TABLE ONLY customerlogin
    ADD CONSTRAINT customerlogin_session_key UNIQUE (session);



ALTER TABLE ONLY entities
    ADD CONSTRAINT entities_pkey PRIMARY KEY (id);



ALTER TABLE ONLY financial_year
    ADD CONSTRAINT financial_year_pkey PRIMARY KEY (id);



ALTER TABLE ONLY hierarchy_paths
    ADD CONSTRAINT hierarchy_paths_pkey PRIMARY KEY (ancestor_id, descendant_id);



ALTER TABLE ONLY lastused
    ADD CONSTRAINT lastused_pkey PRIMARY KEY (id);



ALTER TABLE ONLY org
    ADD CONSTRAINT org_pkey PRIMARY KEY (id);



ALTER TABLE ONLY paymentmethod
    ADD CONSTRAINT paymentmethod_pkey PRIMARY KEY (id);



ALTER TABLE ONLY report
    ADD CONSTRAINT report_pkey PRIMARY KEY (reportid);



ALTER TABLE ONLY schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);



ALTER TABLE ONLY search_irrelevant_words
    ADD CONSTRAINT search_irrelevant_words_pkey PRIMARY KEY (id);



ALTER TABLE ONLY tax
    ADD CONSTRAINT tax_pkey PRIMARY KEY (id);



ALTER TABLE ONLY trf
    ADD CONSTRAINT trf_pkey PRIMARY KEY (id);



ALTER TABLE ONLY users
    ADD CONSTRAINT users_email_key UNIQUE (email);



ALTER TABLE ONLY users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);



ALTER TABLE ONLY vat_settlement
    ADD CONSTRAINT vat_settlement_pkey PRIMARY KEY (id);



ALTER TABLE ONLY vendor
    ADD CONSTRAINT vendor_pkey PRIMARY KEY (id);



ALTER TABLE ONLY yearend
    ADD CONSTRAINT yearend_pkey PRIMARY KEY (id);



CREATE INDEX acc_trans_chart_id_key ON acc_trans USING btree (chart_id);



CREATE INDEX acc_trans_chart_id_transdate_approved_trans_id ON acc_trans USING btree (chart_id, transdate, approved, trans_id, amount);



CREATE INDEX acc_trans_source_key ON acc_trans USING btree (lower(source));



CREATE INDEX acc_trans_trans_id_key ON acc_trans USING btree (trans_id);



CREATE INDEX acc_trans_transdate_key ON acc_trans USING btree (transdate);



CREATE INDEX ap_quonumber_key ON ap USING btree (lower(quonumber));



CREATE INDEX ar_quonumber_key ON ar USING btree (lower(quonumber));



CREATE INDEX assembly_id_key ON assembly USING btree (id);



CREATE INDEX audittrail_trans_id_key ON audittrail USING btree (trans_id);



CREATE INDEX cargo_id_key ON cargo USING btree (id, trans_id);



CREATE UNIQUE INDEX chart_accno_key ON oldchart USING btree (accno);



CREATE INDEX chart_category_key ON oldchart USING btree (category);



CREATE INDEX chart_gifi_accno_key ON oldchart USING btree (gifi_accno);



CREATE INDEX chart_id_key ON oldchart USING btree (id);



CREATE INDEX chart_link_key ON oldchart USING btree (link);



CREATE INDEX customer_contact_key ON customer USING btree (lower((contact)::text));



CREATE INDEX customer_customer_id_key ON customertax USING btree (customer_id);



CREATE INDEX customer_customernumber_key ON customer USING btree (customernumber);



CREATE INDEX customer_name_key ON customer USING btree (lower((name)::text));



CREATE INDEX department_id_key ON department USING btree (id);



CREATE INDEX employee_id_key ON employee USING btree (id);



CREATE UNIQUE INDEX employee_login_key ON employee USING btree (login);



CREATE INDEX employee_name_key ON employee USING btree (name);



CREATE INDEX exchangerate_ct_key ON exchangerate USING btree (curr, transdate);



CREATE INDEX fifo_parts_id ON fifo USING btree (parts_id);



CREATE INDEX fifo_trans_id ON fifo USING btree (trans_id);



CREATE UNIQUE INDEX gifi_accno_key ON gifi USING btree (accno);



CREATE INDEX gl_description_key ON gl USING btree (lower(description));



CREATE INDEX gl_employee_id_key ON gl USING btree (employee_id);



CREATE INDEX gl_id_key ON gl USING btree (id);



CREATE INDEX gl_reference_key ON gl USING btree (reference);



CREATE INDEX gl_transdate_key ON gl USING btree (transdate);



CREATE INDEX idx_entities_deleted ON entities USING btree (deleted_at);



CREATE INDEX idx_entities_org ON entities USING btree (organization_id);



CREATE INDEX idx_entities_org_parent ON entities USING btree (organization_id, parent_id);



CREATE INDEX idx_entities_parent ON entities USING btree (parent_id);



CREATE INDEX idx_entities_type ON entities USING btree (type);



CREATE INDEX idx_hierarchy_ancestor ON hierarchy_paths USING btree (ancestor_id);



CREATE INDEX idx_hierarchy_descendant ON hierarchy_paths USING btree (descendant_id);



CREATE INDEX idx_hierarchy_paths_depth ON hierarchy_paths USING btree (depth);



CREATE INDEX idx_users_deleted ON users USING btree (deleted_at);



CREATE INDEX idx_users_entity ON users USING btree (entity_id);



CREATE INDEX idx_users_last_login ON users USING btree (last_login_at);



CREATE INDEX inventory_invoice_id ON inventory USING btree (invoice_id);



CREATE INDEX inventory_parts_id_key ON inventory USING btree (parts_id);



CREATE INDEX jcitems_id_key ON jcitems USING btree (id);



CREATE UNIQUE INDEX language_code_key ON language USING btree (code);



CREATE INDEX makemodel_make_key ON makemodel USING btree (lower(make));



CREATE INDEX makemodel_model_key ON makemodel USING btree (lower(model));



CREATE INDEX makemodel_parts_id_key ON makemodel USING btree (parts_id);



CREATE INDEX oe_employee_id_key ON oe USING btree (employee_id);



CREATE INDEX oe_id_key ON oe USING btree (id);



CREATE INDEX oe_ordnumber_key ON oe USING btree (ordnumber);



CREATE INDEX oe_transdate_key ON oe USING btree (transdate);



CREATE INDEX orderitems_id_key ON orderitems USING btree (id);



CREATE INDEX orderitems_trans_id_key ON orderitems USING btree (trans_id);



CREATE INDEX parts_description_key ON parts USING btree (lower(description));



CREATE INDEX parts_id_key ON parts USING btree (id);



CREATE INDEX parts_partnumber_key ON parts USING btree (lower(partnumber));



CREATE INDEX partscustomer_customer_id_key ON partscustomer USING btree (customer_id);



CREATE INDEX partscustomer_parts_id_key ON partscustomer USING btree (parts_id);



CREATE INDEX partsgroup_id_key ON partsgroup USING btree (id);



CREATE UNIQUE INDEX partsgroup_key ON partsgroup USING btree (partsgroup);



CREATE INDEX partstax_parts_id_key ON partstax USING btree (parts_id);



CREATE INDEX partsvendor_parts_id_key ON partsvendor USING btree (parts_id);



CREATE INDEX partsvendor_vendor_id_key ON partsvendor USING btree (vendor_id);



CREATE INDEX pricegroup_id_key ON pricegroup USING btree (id);



CREATE INDEX pricegroup_pricegroup_key ON pricegroup USING btree (pricegroup);



CREATE INDEX project_id_key ON project USING btree (id);



CREATE UNIQUE INDEX projectnumber_key ON project USING btree (projectnumber);



CREATE INDEX shipto_trans_id_key ON shipto USING btree (trans_id);



CREATE INDEX status_trans_id_key ON status USING btree (trans_id);



CREATE INDEX translation_trans_id_key ON translation USING btree (trans_id);



CREATE INDEX vendor_contact_key ON vendor USING btree (lower((contact)::text));



CREATE INDEX vendor_name_key ON vendor USING btree (lower((name)::text));



CREATE INDEX vendor_vendornumber_key ON vendor USING btree (vendornumber);



CREATE INDEX vendortax_vendor_id_key ON vendortax USING btree (vendor_id);



CREATE TRIGGER check_department AFTER INSERT OR UPDATE ON ap FOR EACH ROW EXECUTE FUNCTION check_department();



CREATE TRIGGER check_department AFTER INSERT OR UPDATE ON ar FOR EACH ROW EXECUTE FUNCTION check_department();



CREATE TRIGGER check_department AFTER INSERT OR UPDATE ON gl FOR EACH ROW EXECUTE FUNCTION check_department();



CREATE TRIGGER check_department AFTER INSERT OR UPDATE ON oe FOR EACH ROW EXECUTE FUNCTION check_department();



CREATE TRIGGER del_customer AFTER DELETE ON customer FOR EACH ROW EXECUTE FUNCTION del_customer();



CREATE TRIGGER del_recurring AFTER DELETE ON ap FOR EACH ROW EXECUTE FUNCTION del_recurring();



CREATE TRIGGER del_recurring AFTER DELETE ON ar FOR EACH ROW EXECUTE FUNCTION del_recurring();



CREATE TRIGGER del_recurring AFTER DELETE ON gl FOR EACH ROW EXECUTE FUNCTION del_recurring();



CREATE TRIGGER del_recurring AFTER DELETE ON oe FOR EACH ROW EXECUTE FUNCTION del_recurring();



CREATE TRIGGER del_vendor AFTER DELETE ON vendor FOR EACH ROW EXECUTE FUNCTION del_vendor();



CREATE TRIGGER del_yearend AFTER DELETE ON gl FOR EACH ROW EXECUTE FUNCTION del_yearend();



CREATE TRIGGER trg_entity_paths AFTER INSERT ON entities FOR EACH ROW EXECUTE FUNCTION update_entity_paths();



CREATE TRIGGER trg_update_entity_timestamp BEFORE UPDATE ON entities FOR EACH ROW EXECUTE FUNCTION update_timestamps();



CREATE TRIGGER trg_update_org_timestamp BEFORE UPDATE ON org FOR EACH ROW EXECUTE FUNCTION update_timestamps();



CREATE TRIGGER trg_update_user_timestamp BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_timestamps();



ALTER TABLE ONLY blink_import_process
    ADD CONSTRAINT blink_import_process_bank_account_id_fkey FOREIGN KEY (bank_account_id) REFERENCES bank_account(id);



ALTER TABLE ONLY blink_import_process_log
    ADD CONSTRAINT blink_import_process_log_banking_import_event_id_fkey FOREIGN KEY (banking_import_event_id) REFERENCES banking_import_event(id);



ALTER TABLE ONLY blink_import_process_log
    ADD CONSTRAINT blink_import_process_log_blink_import_process_id_fkey FOREIGN KEY (blink_import_process_id) REFERENCES blink_import_process(id);



ALTER TABLE ONLY booking_to_settlement
    ADD CONSTRAINT booking_to_settlement_settlement_id_fkey FOREIGN KEY (settlement_id) REFERENCES vat_settlement(id);



ALTER TABLE ONLY entities
    ADD CONSTRAINT entities_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES org(id) ON DELETE CASCADE;



ALTER TABLE ONLY entities
    ADD CONSTRAINT entities_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES entities(id) ON DELETE CASCADE;



ALTER TABLE ONLY financial_year
    ADD CONSTRAINT fk_yearend FOREIGN KEY (yearend_id) REFERENCES yearend(id) ON DELETE SET NULL;



ALTER TABLE ONLY hierarchy_paths
    ADD CONSTRAINT hierarchy_paths_ancestor_id_fkey FOREIGN KEY (ancestor_id) REFERENCES entities(id) ON DELETE CASCADE;



ALTER TABLE ONLY hierarchy_paths
    ADD CONSTRAINT hierarchy_paths_descendant_id_fkey FOREIGN KEY (descendant_id) REFERENCES entities(id) ON DELETE CASCADE;



ALTER TABLE ONLY users
    ADD CONSTRAINT users_entity_id_fkey FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE RESTRICT;



ALTER TABLE ONLY vr
    ADD CONSTRAINT vr_br_id_fkey FOREIGN KEY (br_id) REFERENCES br(id) ON DELETE CASCADE;



