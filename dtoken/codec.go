package dtoken

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/crypto/gaes"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/grand"
)

// Encoder defines the encoder interface | 编码器接口定义
type Encoder interface {
	Encode(ctx context.Context, userKey string) (token string, err error) // Encode method to generate token | 生成 Token 的编码方法
}

// Decoder defines the decoder interface | 解码器接口定义
type Decoder interface {
	Decrypt(ctx context.Context, token string) (userKey string, err error) // Decrypt method to extract userKey from token | 解码方法，从 Token 中提取用户标识
}

// Codec combines both encoder and decoder interfaces | Codec 组合了编码器和解码器接口
type Codec interface {
	Encoder
	Decoder
}

// DefaultCodec default implementation of Codec | 默认编解码实现
type DefaultCodec struct {
	// Delimiter used to separate userKey and random string | 用于分隔用户标识和随机字符串的分隔符
	Delimiter string
	// EncryptKey for token encryption | Token 加密密钥
	EncryptKey []byte
}

// NewDefaultCodec creates a new DefaultCodec instance | 创建一个新的 DefaultCodec 实例
func NewDefaultCodec(delimiter string, encryptKey []byte) *DefaultCodec {
	return &DefaultCodec{
		Delimiter:  delimiter,
		EncryptKey: encryptKey,
	}
}

// Encode method to generate token | 编码方法生成 Token
func (c *DefaultCodec) Encode(ctx context.Context, userKey string) (token string, err error) {
	if userKey == "" {
		return "", errors.New(MsgErrUserKeyEmpty) // Error when userKey is empty | 用户标识为空时返回错误
	}
	// Generate a random string | 生成一个随机字符串
	randStr, err := gmd5.Encrypt(grand.Letters(10))
	if err != nil {
		return "", err
	}
	encryptBeforeStr := userKey + c.Delimiter + randStr
	// Encrypt the combined string | 加密拼接后的字符串
	encryptByte, err := gaes.Encrypt([]byte(encryptBeforeStr), c.EncryptKey)
	if err != nil {
		return "", err
	}
	// Return base64 encoded token | 返回 Base64 编码后的 Token
	return gbase64.EncodeToString(encryptByte), nil
}

// Decrypt method to extract userKey from token | 解码方法从 Token 中提取用户标识
func (c *DefaultCodec) Decrypt(ctx context.Context, token string) (userKey string, err error) {
	if token == "" {
		return "", errors.New(MsgErrTokenEmpty) // Error when token is empty | Token 为空时返回错误
	}
	// Decode the base64 token | 解码 Base64 编码的 Token
	token64, err := gbase64.Decode([]byte(token))
	if err != nil {
		return "", err
	}
	// Decrypt the decoded token | 解密解码后的 Token
	decryptStr, err := gaes.Decrypt(token64, c.EncryptKey)
	if err != nil {
		return "", err
	}
	// Split the decrypted string to extract userKey | 分割解密后的字符串，提取用户标识
	decryptArray := gstr.Split(string(decryptStr), c.Delimiter)
	if len(decryptArray) < 2 {
		return "", errors.New(MsgErrTokenLen) // Error when the token length is invalid | Token 长度无效时返回错误
	}
	return decryptArray[0], nil
}
