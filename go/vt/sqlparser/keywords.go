/*
Copyright 2021 The Vitess Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sqlparser

import (
	"fmt"
	"strings"
)

type keyword struct {
	name string
	id   int
}

func (k *keyword) match(input []byte) bool {
	if len(input) != len(k.name) {
		return false
	}
	for i, c := range input {
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		if k.name[i] != c {
			return false
		}
	}
	return true
}

func (k *keyword) matchStr(input string) bool {
	return keywordASCIIMatch(input, k.name)
}

func keywordASCIIMatch(input string, expected string) bool {
	if len(input) != len(expected) {
		return false
	}
	for i := 0; i < len(input); i++ {
		c := input[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		if expected[i] != c {
			return false
		}
	}
	return true
}

// keywords is a table of mysql keywords that fall into two categories:
// 1) keywords considered reserved by MySQL
// 2) keywords for us to handle specially in sql.y
//
// Those marked as UNUSED are likely reserved keywords. We add them here so that
// when rewriting queries we can properly backtick quote them so they don't cause issues
//
// NOTE: If you add new keywords, add them also to the reserved_keywords or
// non_reserved_keywords grammar in sql.y -- this will allow the keyword to be used
// in identifiers. See the docs for each grammar to determine which one to put it into.
var keywords = []keyword{
	{"_armscii8", UNDERSCORE_ARMSCII8},
	{"_ascii", UNDERSCORE_ASCII},
	{"_big5", UNDERSCORE_BIG5},
	{"_binary", UNDERSCORE_BINARY},
	{"_cp1250", UNDERSCORE_CP1250},
	{"_cp1251", UNDERSCORE_CP1251},
	{"_cp1256", UNDERSCORE_CP1256},
	{"_cp1257", UNDERSCORE_CP1257},
	{"_cp850", UNDERSCORE_CP850},
	{"_cp852", UNDERSCORE_CP852},
	{"_cp866", UNDERSCORE_CP866},
	{"_cp932", UNDERSCORE_CP932},
	{"_dec8", UNDERSCORE_DEC8},
	{"_eucjpms", UNDERSCORE_EUCJPMS},
	{"_euckr", UNDERSCORE_EUCKR},
	{"_gb18030", UNDERSCORE_GB18030},
	{"_gb2312", UNDERSCORE_GB2312},
	{"_gbk", UNDERSCORE_GBK},
	{"_geostd8", UNDERSCORE_GEOSTD8},
	{"_greek", UNDERSCORE_GREEK},
	{"_hebrew", UNDERSCORE_HEBREW},
	{"_hp8", UNDERSCORE_HP8},
	{"_keybcs2", UNDERSCORE_KEYBCS2},
	{"_koi8r", UNDERSCORE_KOI8R},
	{"_koi8u", UNDERSCORE_KOI8U},
	{"_latin1", UNDERSCORE_LATIN1},
	{"_latin2", UNDERSCORE_LATIN2},
	{"_latin5", UNDERSCORE_LATIN5},
	{"_latin7", UNDERSCORE_LATIN7},
	{"_macce", UNDERSCORE_MACCE},
	{"_macroman", UNDERSCORE_MACROMAN},
	{"_sjis", UNDERSCORE_SJIS},
	{"_swe7", UNDERSCORE_SWE7},
	{"_tis620", UNDERSCORE_TIS620},
	{"_ucs2", UNDERSCORE_UCS2},
	{"_ujis", UNDERSCORE_UJIS},
	{"_utf16", UNDERSCORE_UTF16},
	{"_utf16le", UNDERSCORE_UTF16LE},
	{"_utf32", UNDERSCORE_UTF32},
	{"_utf8", UNDERSCORE_UTF8},
	{"_utf8mb4", UNDERSCORE_UTF8MB4},
	{"accessible", UNUSED},
	{"action", ACTION},
	{"add", ADD},
	{"after", AFTER},
	{"against", AGAINST},
	{"algorithm", ALGORITHM},
	{"all", ALL},
	{"alter", ALTER},
	{"always", ALWAYS},
	{"analyze", ANALYZE},
	{"and", AND},
	{"as", AS},
	{"asc", ASC},
	{"ascii", ASCII},
	{"asensitive", UNUSED},
	{"auto_increment", AUTO_INCREMENT},
	{"avg_row_length", AVG_ROW_LENGTH},
	{"before", UNUSED},
	{"begin", BEGIN},
	{"between", BETWEEN},
	{"bigint", BIGINT},
	{"binary", BINARY},
	{"bit", BIT},
	{"blob", BLOB},
	{"bool", BOOL},
	{"boolean", BOOLEAN},
	{"both", UNUSED},
	{"by", BY},
	{"call", CALL},
	{"cancel", CANCEL},
	{"cascade", CASCADE},
	{"cascaded", CASCADED},
	{"case", CASE},
	{"cast", CAST},
	{"channel", CHANNEL},
	{"change", CHANGE},
	{"char", CHAR},
	{"character", CHARACTER},
	{"charset", CHARSET},
	{"check", CHECK},
	{"checksum", CHECKSUM},
	{"cleanup", CLEANUP},
	{"coalesce", COALESCE},
	{"code", CODE},
	{"collate", COLLATE},
	{"collation", COLLATION},
	{"column", COLUMN},
	{"columns", COLUMNS},
	{"comment", COMMENT_KEYWORD},
	{"committed", COMMITTED},
	{"commit", COMMIT},
	{"compact", COMPACT},
	{"complete", COMPLETE},
	{"compressed", COMPRESSED},
	{"compression", COMPRESSION},
	{"condition", UNUSED},
	{"connection", CONNECTION},
	{"constraint", CONSTRAINT},
	{"continue", UNUSED},
	{"convert", CONVERT},
	{"copy", COPY},
	{"cume_dist", UNUSED},
	{"substr", SUBSTRING},
	{"subpartition", SUBPARTITION},
	{"subpartitions", SUBPARTITIONS},
	{"substring", SUBSTRING},
	{"create", CREATE},
	{"cross", CROSS},
	{"csv", CSV},
	{"current_date", CURRENT_DATE},
	{"current_time", CURRENT_TIME},
	{"current_timestamp", CURRENT_TIMESTAMP},
	{"current_user", CURRENT_USER},
	{"cursor", UNUSED},
	{"data", DATA},
	{"database", DATABASE},
	{"databases", DATABASES},
	{"day", DAY},
	{"day_hour", DAY_HOUR},
	{"day_microsecond", DAY_MICROSECOND},
	{"day_minute", DAY_MINUTE},
	{"day_second", DAY_SECOND},
	{"date", DATE},
	{"datetime", DATETIME},
	{"dec", UNUSED},
	{"decimal", DECIMAL_TYPE},
	{"declare", UNUSED},
	{"default", DEFAULT},
	{"definer", DEFINER},
	{"delay_key_write", DELAY_KEY_WRITE},
	{"delayed", UNUSED},
	{"delete", DELETE},
	{"dense_rank", UNUSED},
	{"desc", DESC},
	{"describe", DESCRIBE},
	{"deterministic", UNUSED},
	{"directory", DIRECTORY},
	{"disable", DISABLE},
	{"discard", DISCARD},
	{"disk", DISK},
	{"distinct", DISTINCT},
	{"distinctrow", DISTINCTROW},
	{"div", DIV},
	{"double", DOUBLE},
	{"do", DO},
	{"drop", DROP},
	{"dumpfile", DUMPFILE},
	{"duplicate", DUPLICATE},
	{"dynamic", DYNAMIC},
	{"each", UNUSED},
	{"else", ELSE},
	{"elseif", UNUSED},
	{"empty", UNUSED},
	{"enable", ENABLE},
	{"enclosed", ENCLOSED},
	{"encryption", ENCRYPTION},
	{"end", END},
	{"enforced", ENFORCED},
	{"engine", ENGINE},
	{"engines", ENGINES},
	{"enum", ENUM},
	{"error", ERROR},
	{"escape", ESCAPE},
	{"escaped", ESCAPED},
	{"event", EVENT},
	{"exchange", EXCHANGE},
	{"exclusive", EXCLUSIVE},
	{"execute", EXECUTE},
	{"exists", EXISTS},
	{"exit", UNUSED},
	{"explain", EXPLAIN},
	{"expansion", EXPANSION},
	{"export", EXPORT},
	{"extended", EXTENDED},
	{"extract", EXTRACT},
	{"false", FALSE},
	{"fetch", UNUSED},
	{"fields", FIELDS},
	{"first", FIRST},
	{"first_value", UNUSED},
	{"fixed", FIXED},
	{"float", FLOAT_TYPE},
	{"float4", UNUSED},
	{"float8", UNUSED},
	{"flush", FLUSH},
	{"for", FOR},
	{"force", FORCE},
	{"foreign", FOREIGN},
	{"format", FORMAT},
	{"from", FROM},
	{"full", FULL},
	{"fulltext", FULLTEXT},
	{"function", FUNCTION},
	{"general", GENERAL},
	{"generated", GENERATED},
	{"geometry", GEOMETRY},
	{"geometrycollection", GEOMETRYCOLLECTION},
	{"get", UNUSED},
	{"global", GLOBAL},
	{"gtid_executed", GTID_EXECUTED},
	{"grant", UNUSED},
	{"group", GROUP},
	{"grouping", UNUSED},
	{"groups", UNUSED},
	{"group_concat", GROUP_CONCAT},
	{"hash", HASH},
	{"having", HAVING},
	{"header", HEADER},
	{"high_priority", UNUSED},
	{"hosts", HOSTS},
	{"hour", HOUR},
	{"hour_microsecond", HOUR_MICROSECOND},
	{"hour_minute", HOUR_MINUTE},
	{"hour_second", HOUR_SECOND},
	{"if", IF},
	{"ignore", IGNORE},
	{"import", IMPORT},
	{"in", IN},
	{"index", INDEX},
	{"indexes", INDEXES},
	{"infile", UNUSED},
	{"inout", UNUSED},
	{"inner", INNER},
	{"inplace", INPLACE},
	{"insensitive", UNUSED},
	{"insert", INSERT},
	{"insert_method", INSERT_METHOD},
	{"int", INT},
	{"int1", UNUSED},
	{"int2", UNUSED},
	{"int3", UNUSED},
	{"int4", UNUSED},
	{"int8", UNUSED},
	{"integer", INTEGER},
	{"interval", INTERVAL},
	{"into", INTO},
	{"io_after_gtids", UNUSED},
	{"is", IS},
	{"isolation", ISOLATION},
	{"iterate", UNUSED},
	{"invoker", INVOKER},
	{"join", JOIN},
	{"json", JSON},
	{"json_table", UNUSED},
	{"key", KEY},
	{"keys", KEYS},
	{"keyspaces", KEYSPACES},
	{"key_block_size", KEY_BLOCK_SIZE},
	{"kill", UNUSED},
	{"lag", UNUSED},
	{"language", LANGUAGE},
	{"last", LAST},
	{"last_value", UNUSED},
	{"last_insert_id", LAST_INSERT_ID},
	{"lateral", UNUSED},
	{"lead", UNUSED},
	{"leading", UNUSED},
	{"leave", UNUSED},
	{"left", LEFT},
	{"less", LESS},
	{"level", LEVEL},
	{"like", LIKE},
	{"limit", LIMIT},
	{"linear", LINEAR},
	{"lines", LINES},
	{"linestring", LINESTRING},
	{"list", LIST},
	{"load", LOAD},
	{"local", LOCAL},
	{"localtime", LOCALTIME},
	{"localtimestamp", LOCALTIMESTAMP},
	{"lock", LOCK},
	{"logs", LOGS},
	{"long", UNUSED},
	{"longblob", LONGBLOB},
	{"longtext", LONGTEXT},
	{"loop", UNUSED},
	{"low_priority", LOW_PRIORITY},
	{"manifest", MANIFEST},
	{"master_bind", UNUSED},
	{"match", MATCH},
	{"max_rows", MAX_ROWS},
	{"maxvalue", MAXVALUE},
	{"mediumblob", MEDIUMBLOB},
	{"mediumint", MEDIUMINT},
	{"mediumtext", MEDIUMTEXT},
	{"memory", MEMORY},
	{"member", MEMBER},
	{"merge", MERGE},
	{"microsecond", MICROSECOND},
	{"middleint", UNUSED},
	{"min_rows", MIN_ROWS},
	{"minute", MINUTE},
	{"minute_microsecond", MINUTE_MICROSECOND},
	{"minute_second", MINUTE_SECOND},
	{"mod", MOD},
	{"mode", MODE},
	{"modify", MODIFY},
	{"modifies", UNUSED},
	{"multilinestring", MULTILINESTRING},
	{"multipoint", MULTIPOINT},
	{"multipolygon", MULTIPOLYGON},
	{"month", MONTH},
	{"name", NAME},
	{"names", NAMES},
	{"natural", NATURAL},
	{"nchar", NCHAR},
	{"next", NEXT},
	{"no", NO},
	{"none", NONE},
	{"not", NOT},
	{"no_write_to_binlog", NO_WRITE_TO_BINLOG},
	{"nth_value", UNUSED},
	{"ntile", UNUSED},
	{"null", NULL},
	{"numeric", NUMERIC},
	{"of", UNUSED},
	{"off", OFF},
	{"offset", OFFSET},
	{"on", ON},
	{"only", ONLY},
	{"open", OPEN},
	{"optimize", OPTIMIZE},
	{"optimizer_costs", OPTIMIZER_COSTS},
	{"option", OPTION},
	{"optionally", OPTIONALLY},
	{"or", OR},
	{"order", ORDER},
	{"out", UNUSED},
	{"outer", OUTER},
	{"outfile", OUTFILE},
	{"over", UNUSED},
	{"overwrite", OVERWRITE},
	{"pack_keys", PACK_KEYS},
	{"parser", PARSER},
	{"partition", PARTITION},
	{"partitions", PARTITIONS},
	{"partitioning", PARTITIONING},
	{"password", PASSWORD},
	{"percent_rank", UNUSED},
	{"plugins", PLUGINS},
	{"point", POINT},
	{"polygon", POLYGON},
	{"precision", UNUSED},
	{"prepare", PREPARE},
	{"primary", PRIMARY},
	{"privileges", PRIVILEGES},
	{"processlist", PROCESSLIST},
	{"procedure", PROCEDURE},
	{"query", QUERY},
	{"range", RANGE},
	{"quarter", QUARTER},
	{"rank", UNUSED},
	{"read", READ},
	{"reads", UNUSED},
	{"read_write", UNUSED},
	{"real", REAL},
	{"rebuild", REBUILD},
	{"recursive", RECURSIVE},
	{"redundant", REDUNDANT},
	{"references", REFERENCES},
	{"regexp", REGEXP},
	{"relay", RELAY},
	{"release", RELEASE},
	{"remove", REMOVE},
	{"rename", RENAME},
	{"reorganize", REORGANIZE},
	{"repair", REPAIR},
	{"repeat", UNUSED},
	{"repeatable", REPEATABLE},
	{"replace", REPLACE},
	{"require", UNUSED},
	{"resignal", UNUSED},
	{"restrict", RESTRICT},
	{"return", UNUSED},
	{"retry", RETRY},
	{"revert", REVERT},
	{"revoke", UNUSED},
	{"right", RIGHT},
	{"rlike", REGEXP},
	{"rollback", ROLLBACK},
	{"row", UNUSED},
	{"row_format", ROW_FORMAT},
	{"row_number", UNUSED},
	{"rows", UNUSED},
	{"s3", S3},
	{"savepoint", SAVEPOINT},
	{"schema", SCHEMA},
	{"schemas", SCHEMAS},
	{"second", SECOND},
	{"second_microsecond", SECOND_MICROSECOND},
	{"security", SECURITY},
	{"select", SELECT},
	{"sensitive", UNUSED},
	{"separator", SEPARATOR},
	{"sequence", SEQUENCE},
	{"serializable", SERIALIZABLE},
	{"session", SESSION},
	{"set", SET},
	{"share", SHARE},
	{"shared", SHARED},
	{"show", SHOW},
	{"signal", UNUSED},
	{"signed", SIGNED},
	{"slow", SLOW},
	{"smallint", SMALLINT},
	{"spatial", SPATIAL},
	{"specific", UNUSED},
	{"sql", SQL},
	{"sqlexception", UNUSED},
	{"sqlstate", UNUSED},
	{"sqlwarning", UNUSED},
	{"sql_big_result", UNUSED},
	{"sql_cache", SQL_CACHE},
	{"sql_calc_found_rows", SQL_CALC_FOUND_ROWS},
	{"sql_no_cache", SQL_NO_CACHE},
	{"sql_small_result", UNUSED},
	{"ssl", UNUSED},
	{"start", START},
	{"starting", STARTING},
	{"stats_auto_recalc", STATS_AUTO_RECALC},
	{"stats_persistent", STATS_PERSISTENT},
	{"stats_sample_pages", STATS_SAMPLE_PAGES},
	{"status", STATUS},
	{"storage", STORAGE},
	{"stored", STORED},
	{"straight_join", STRAIGHT_JOIN},
	{"stream", STREAM},
	{"system", UNUSED},
	{"table", TABLE},
	{"tables", TABLES},
	{"tablespace", TABLESPACE},
	{"temporary", TEMPORARY},
	{"temptable", TEMPTABLE},
	{"terminated", TERMINATED},
	{"text", TEXT},
	{"than", THAN},
	{"then", THEN},
	{"time", TIME},
	{"timestamp", TIMESTAMP},
	{"timestampadd", TIMESTAMPADD},
	{"timestampdiff", TIMESTAMPDIFF},
	{"tinyblob", TINYBLOB},
	{"tinyint", TINYINT},
	{"tinytext", TINYTEXT},
	{"to", TO},
	{"trailing", UNUSED},
	{"transaction", TRANSACTION},
	{"tree", TREE},
	{"traditional", TRADITIONAL},
	{"trigger", TRIGGER},
	{"triggers", TRIGGERS},
	{"true", TRUE},
	{"truncate", TRUNCATE},
	{"uncommitted", UNCOMMITTED},
	{"undefined", UNDEFINED},
	{"undo", UNUSED},
	{"unicode", UNICODE},
	{"union", UNION},
	{"unique", UNIQUE},
	{"unlock", UNLOCK},
	{"unsigned", UNSIGNED},
	{"update", UPDATE},
	{"upgrade", UPGRADE},
	{"usage", UNUSED},
	{"use", USE},
	{"user", USER},
	{"user_resources", USER_RESOURCES},
	{"using", USING},
	{"utc_date", UTC_DATE},
	{"utc_time", UTC_TIME},
	{"utc_timestamp", UTC_TIMESTAMP},
	{"validation", VALIDATION},
	{"values", VALUES},
	{"variables", VARIABLES},
	{"varbinary", VARBINARY},
	{"varchar", VARCHAR},
	{"varcharacter", UNUSED},
	{"varying", UNUSED},
	{"vgtid_executed", VGTID_EXECUTED},
	{"virtual", VIRTUAL},
	{"vindex", VINDEX},
	{"vindexes", VINDEXES},
	{"view", VIEW},
	{"vitess", VITESS},
	{"vitess_keyspaces", VITESS_KEYSPACES},
	{"vitess_metadata", VITESS_METADATA},
	{"vitess_migration", VITESS_MIGRATION},
	{"vitess_migrations", VITESS_MIGRATIONS},
	{"vitess_replication_status", VITESS_REPLICATION_STATUS},
	{"vitess_shards", VITESS_SHARDS},
	{"vitess_tablets", VITESS_TABLETS},
	{"vschema", VSCHEMA},
	{"vstream", VSTREAM},
	{"warnings", WARNINGS},
	{"weight_string", WEIGHT_STRING},
	{"when", WHEN},
	{"where", WHERE},
	{"while", UNUSED},
	{"window", UNUSED},
	{"with", WITH},
	{"without", WITHOUT},
	{"work", WORK},
	{"write", WRITE},
	{"xor", XOR},
	{"year", YEAR},
	{"year_month", YEAR_MONTH},
	{"zerofill", ZEROFILL},
}

// keywordStrings contains the reverse mapping of token to keyword strings
var keywordStrings = map[int]string{}

// keywordLookupTable is a perfect hash map that maps **case insensitive** keyword names to their ids
var keywordLookupTable *caseInsensitiveTable

type caseInsensitiveTable struct {
	h map[uint64]keyword
}

func buildCaseInsensitiveTable(keywords []keyword) *caseInsensitiveTable {
	table := &caseInsensitiveTable{
		h: make(map[uint64]keyword, len(keywords)),
	}

	for _, kw := range keywords {
		hash := fnv1aIstr(offset64, kw.name)
		if _, exists := table.h[hash]; exists {
			panic("collision in caseInsensitiveTable")
		}
		table.h[hash] = kw
	}
	return table
}

func (cit *caseInsensitiveTable) LookupString(name string) (int, bool) {
	hash := fnv1aIstr(offset64, name)
	if candidate, ok := cit.h[hash]; ok {
		return candidate.id, candidate.matchStr(name)
	}
	return 0, false
}

func (cit *caseInsensitiveTable) Lookup(name []byte) (int, bool) {
	hash := fnv1aI(offset64, name)
	if candidate, ok := cit.h[hash]; ok {
		return candidate.id, candidate.match(name)
	}
	return 0, false
}

func init() {
	for _, kw := range keywords {
		if kw.id == UNUSED {
			continue
		}
		if kw.name != strings.ToLower(kw.name) {
			panic(fmt.Sprintf("keyword %q must be lowercase in table", kw.name))
		}
		keywordStrings[kw.id] = kw.name
	}

	keywordLookupTable = buildCaseInsensitiveTable(keywords)
}

// KeywordString returns the string corresponding to the given keyword
func KeywordString(id int) string {
	str, ok := keywordStrings[id]
	if !ok {
		return ""
	}
	return str
}

const offset64 = uint64(14695981039346656037)
const prime64 = uint64(1099511628211)

func fnv1aI(h uint64, s []byte) uint64 {
	for _, c := range s {
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		h = (h ^ uint64(c)) * prime64
	}
	return h
}

func fnv1aIstr(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		h = (h ^ uint64(c)) * prime64
	}
	return h
}
