# Go 1.23

## strucrts

構造体のメモリレイアウトなどのプロパティを設定するための型を提供する。

今回は `HostLayout` のみの追加。cgo などを使って他の言語とやりとりする際にメモリレイアウトが異なると GC がうまく機能しないらしい。
詳しくは、[提案時の PR](https://github.com/golang/go/issues/66408) や、[このページ](https://medium.com/@ksandeeptech07/golang-1-23-new-iterators-and-struct-packages-ead9a0d3e92f)の HostLayout の項目を参照してください。

## archive/tar

`FileInfoHeader` に `FileInfoNames` を実装した引数を渡すことで、`Header.Uname` と `Header.Gname` を書き換えられる。
呼び出し元は Uname と Gname を直接指定することで、システム依存の名前検索を避けることができる。

[Go Playground](https://go.dev/play/p/MGVhPKbrJ_4)

## crypto/tls

ClientHello の暗号化（ECH: Encrypted Client Hello）にクライアントのみ対応(#63369)。
サーバ側は 1.24 に対応予定（#68500）らしいが、現在は Proposal 段階。

Server はまだ対応していない模様

ECH について詳しくは
- [Cloudflare](https://blog.cloudflare.com/ja-jp/announcing-encrypted-client-hello-ja-jp)
- [Internet-Draft](https://www.ietf.org/archive/id/draft-ietf-tls-esni-18.html)
- [ECH を BoringSSL で試してみる](https://asnokaze.hatenablog.com/entry/2021/07/23/224703)

試し方

```bash
# launch server supporting ECH using boringssl
git clone --branch=master --depth=1 https://boringssl..googlesource.com/boringssl
cd boringssl
cmake -B build
make -C build
cd ../
./boringssl/build/tool/bssl generate-ech  \
   -out-ech-config-list ech_config_list.data \
   -out-ech-config ech_config.data \
   -out-private-key ech.key \
   -public-name public.example.com \
   -config-id 0
./boringssl/build/tool/bssl server -accept 4430 \
    -ech-key ech.key \ 
    -ech-config ech_config.data &

# execute client with ECH
go install golang.org/dl/gotip@latest
gotip download
gotip run ./cmd/tls \
    -ech-config-list echo_config_list.data \
    -server-name private.example.org \
    -addr localhost:4430

# Server 側の出力に EncryptedClientHello: yes が出力される
```

QUIC の状態変化のイベント(`QUIC{Resume,Store}Session`)と、セッションチケットにデータを追加する方法を追加。
これにより、TLS1.3 の 0-RTT と QUIC をより簡単に統合できる。 #63691

- [参考](https://eng-blog.iij.ad.jp/archives/10620)

クライアンとが対応可能な暗号スイート(CipherSuite フィールド)のデフォルト値から 3DES が削除された。

- [Go Playground](https://go.dev/play/p/t9JSWvxYeu3?v=gotip)

[ポスト量子暗号](https://ja.wikipedia.org/wiki/%E3%83%9D%E3%82%B9%E3%83%88%E9%87%8F%E5%AD%90%E6%9A%97%E5%8F%B7)である `X25519Kyber768Draft00` を `Config.CurvePreferences` のデフォルト値として有効化。
 
- [Go Playground](https://go.dev/play/p/YZp7rT0Jke9?v=gotip)
- [参考](https://blog.cloudflare.com/post-quantum-cryptography-ga-ja-jp)

X509KeyPair と LoadX509KeyPair で、パースされた証明書が、戻り値の `Certificate.Leaf` に格納されるようになった。

- [Go Playground](https://go.dev/play/p/g94DYbQzrFH?v=gotip)

## crypto/x509

`CreateCertificateRequest` が適切に RSA-PSS 署名アルゴリズムをサポートするようになった。

- [Go Playground](https://go.dev/play/p/TGNgUYvNH5o?v=gotip)

`GODEBUG` 環境変数から `x509sha1=1` のサポートが Go1.24 で外される。これで、Go は SHA1 をサポートしなくなる。

ASN.1 形式の OID 文字列をパースする `ParseOID` を追加。これに伴い、`encoding.{Text,Binary}{Un,}Marshaler` を実装するようになった。

## database/sql

`driver.Valuer` が返した `error` を `DB.{Query,Exec,QueryRow}` などがラップして返すようになりました。
`fmt.Println(err)` などで返した文言は変わらないが、 `errors.Is` で判定ができるようになったということ。

```go

package main

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var errValuer = fmt.Errorf("valuer error")

type StringValuer string

func (s StringValuer) Value() (driver.Value, error) {
	return nil, errValuer
}

func main() {
	os.Remove("./foo.db")

	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
	create table foo (id integer not null primary key, name text);
	delete from foo;
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	_, err = tx.Exec("insert into foo(id, name) values(?, ?)", 1, StringValuer("Hello, World"))
	fmt.Println("err is errValuer ?", errors.Is(err, errValuer))
	// Go1.22: false
	// Go1.23: true
}
```

## debug/elf

OpenBSD で BTCFI を無効化するための `PT_OPENBSD_NOBTCFI` を追加。
詳しくは #66054

> [!note]  CFI とは？
>
> CFI (Control Flow Integrity) はプラグラムの実行中に予期せぬ命令のジャンプが起きないようにする技術です。
> これにより、任意の命令を実行させる攻撃に対するセキュリティが向上します。
>
> ジャンプには、固定されたアドレスに飛ぶ直接的なジャンプ（goto や関数呼び出しなど）と、実行時にアドレスが決まる間接的なジャンプがあります。
> 直接的なジャンプは予測可能で安全ですが、間接的なジャンプは、実行時にアドレスが決まるので攻撃可能です。
> そのため、ここでは間接的なジャンプが対象になります。
>
> また、CFI は IBT (Indirect Branch Tracking) と BTI (Branch Target Identification) があります。
> IBT は、動的にジャンプのターゲットを追跡し、不正なジャンプを防ぎます。
> BTI は、静的にジャンプのターゲットをリスト化し、不正なジャンプを防ぎます。

- [Control-flow integrity](https://en.wikipedia.org/wiki/Control-flow_integrity)
- [Control-flow integrity (OpenBSD)](https://isopenbsdsecu.re/mitigations/forward_edge_cfi/)

ELF ファイルのシンボルタイプ追加

## encoding/binary

`Read/Write` に相当する `[]byte` 用の `Decode/Encode` 関数追加。

- [Go Playground](https://go.dev/play/p/tN8Xse-xKiE?v=gotip)

`[]byte` に Encode した結果を追記できる `Append` を追加。

- [Go Playground](https://go.dev/play/p/dC5JFyrPPz_G?v=gotip)

## go/ast

構文解析木の全ノードに対する iterator を作成する `Preorder` 関数を追加

- [Go Playground](https://go.dev/play/p/QURoCexQdLr?v=gotip)

## go/types

`types.Func` に関数の型(シグネチャ)を返す関数が追加。

- [Go Playground](https://go.dev/play/p/LLbiIxSdkU-?v=gotip)

`types.Alias` に `Rhs` メソッドを追加して、エイリアスのベースの型がわかるようになった。
ジェネリック型のエイリアスのための布石らしい(66559)

- [Go Playground](https://go.dev/play/p/0HLMR4Hz8kN?v=gotip)

同様にいくつか `types.Alias` にメソッドが追加。

また、エイリアス型は、1.22 まで `types.Named` 型として作られていたが、1.23 から `types.Alias` 型として作られる。

- [Go Playground](https://go.dev/play/p/LTtTOXZaBkh?v=gotip)

この挙動は `GODEBUG=gotypesalias=0` とすることで 1.22 以前に戻る。

## math/rand/v2

1.22 で忘れていた `rand.Uint` と `rand.Rand.Uint` が追加。(`UintN` とかは 1.22 で追加されている)

`ChaCha8` が `io.Reader` を実装。


## net

Cookie の value が double-quote (`"`) で囲まれた時の挙動が変わる。

- `fmt.Print(Cookie)` した時、1.23 以降では囲いを維持 (1.22 までは削除されていた)
- `Cookie.Value` は変わらない
- 空文字はどちらも空
- 元の値が double-quote で囲まれていたか判別する `Quoted` フィールド追加

[Go Playground](https://go.dev/play/p/5QWjk88KmRy?v=gotip)

(#46443) そもそも cookie の BNF が RFC と違っていたっぽい。

同一 Name の全ての Cookie を取り出す `Request.NamedCookie` が追加

[Go Playground](https://go.dev/play/p/mGD6qe_QK8b?v=gotip)

Cookie の Partitioned 属性のサポート。

この辺の話: [Google Chromeが勧めるプライバシーサンドボックス技術のひとつ、CHIPSってなんぞ](https://qiita.com/rana_kualu/items/8d8e8cd0d4737e7884e3)

`ServeMux` の `Handle{Func,}` に指定する `pattern` でメソッドとパス名の間の空白(`GET<ここ>/path`)が柔軟になった。

[Go Playground](https://go.dev/play/p/UzyaIDBiIt5?v=gotip)

Cookie のパース関数追加

[Go Playground](https://go.dev/play/p/j-ZqRiKjVGw?v=gotip)

`Serve{Content,File{,FS}}` がエラー応答時に不要なヘッダ（Cache-Control, Content-Encoding, Etag, and Last-Modified ）を削除するようになった。
Content-Encoding を設定するミドルウェアなどを使っている場合は要注意。on-the-fly の圧縮は [Transfer-Encoding](https://developer.mozilla.org/ja/docs/Web/HTTP/Headers/Transfer-Encoding) を使うように変更して。

[Go Playground](https://go.dev/play/p/bC_YRC6J-P4?v=gotip)

入力されたリクエストに対して、`Request.Pattern` にヒットした `ServeMux` のパターンが記載される。

[Go Playground](https://go.dev/play/p/D_ye4VK3tMu?v=gotip)

## net/http/httptest

`NewRequestWithContext` の追加。
