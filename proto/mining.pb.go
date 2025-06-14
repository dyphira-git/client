// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v3.12.4
// source: proto/mining.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Request for a block template
type BlockTemplateRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	MinerAddress  string                 `protobuf:"bytes,1,opt,name=miner_address,json=minerAddress,proto3" json:"miner_address,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BlockTemplateRequest) Reset() {
	*x = BlockTemplateRequest{}
	mi := &file_proto_mining_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BlockTemplateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockTemplateRequest) ProtoMessage() {}

func (x *BlockTemplateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_mining_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockTemplateRequest.ProtoReflect.Descriptor instead.
func (*BlockTemplateRequest) Descriptor() ([]byte, []int) {
	return file_proto_mining_proto_rawDescGZIP(), []int{0}
}

func (x *BlockTemplateRequest) GetMinerAddress() string {
	if x != nil {
		return x.MinerAddress
	}
	return ""
}

// Block data structure
type Block struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Timestamp     int64                  `protobuf:"varint,1,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	PrevBlockHash []byte                 `protobuf:"bytes,2,opt,name=prev_block_hash,json=prevBlockHash,proto3" json:"prev_block_hash,omitempty"`
	Height        int32                  `protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
	Transactions  []*Transaction         `protobuf:"bytes,4,rep,name=transactions,proto3" json:"transactions,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Block) Reset() {
	*x = Block{}
	mi := &file_proto_mining_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Block) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Block) ProtoMessage() {}

func (x *Block) ProtoReflect() protoreflect.Message {
	mi := &file_proto_mining_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Block.ProtoReflect.Descriptor instead.
func (*Block) Descriptor() ([]byte, []int) {
	return file_proto_mining_proto_rawDescGZIP(), []int{1}
}

func (x *Block) GetTimestamp() int64 {
	if x != nil {
		return x.Timestamp
	}
	return 0
}

func (x *Block) GetPrevBlockHash() []byte {
	if x != nil {
		return x.PrevBlockHash
	}
	return nil
}

func (x *Block) GetHeight() int32 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *Block) GetTransactions() []*Transaction {
	if x != nil {
		return x.Transactions
	}
	return nil
}

// Response containing a block template
type BlockTemplateResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Block         *Block                 `protobuf:"bytes,1,opt,name=block,proto3" json:"block,omitempty"`
	Difficulty    int32                  `protobuf:"varint,2,opt,name=difficulty,proto3" json:"difficulty,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BlockTemplateResponse) Reset() {
	*x = BlockTemplateResponse{}
	mi := &file_proto_mining_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BlockTemplateResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockTemplateResponse) ProtoMessage() {}

func (x *BlockTemplateResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_mining_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockTemplateResponse.ProtoReflect.Descriptor instead.
func (*BlockTemplateResponse) Descriptor() ([]byte, []int) {
	return file_proto_mining_proto_rawDescGZIP(), []int{2}
}

func (x *BlockTemplateResponse) GetBlock() *Block {
	if x != nil {
		return x.Block
	}
	return nil
}

func (x *BlockTemplateResponse) GetDifficulty() int32 {
	if x != nil {
		return x.Difficulty
	}
	return 0
}

// Request to submit a mined block
type SubmitBlockRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Block         *Block                 `protobuf:"bytes,1,opt,name=block,proto3" json:"block,omitempty"`
	BlockHash     []byte                 `protobuf:"bytes,2,opt,name=block_hash,json=blockHash,proto3" json:"block_hash,omitempty"`
	Nonce         int32                  `protobuf:"varint,3,opt,name=nonce,proto3" json:"nonce,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubmitBlockRequest) Reset() {
	*x = SubmitBlockRequest{}
	mi := &file_proto_mining_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubmitBlockRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmitBlockRequest) ProtoMessage() {}

func (x *SubmitBlockRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_mining_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubmitBlockRequest.ProtoReflect.Descriptor instead.
func (*SubmitBlockRequest) Descriptor() ([]byte, []int) {
	return file_proto_mining_proto_rawDescGZIP(), []int{3}
}

func (x *SubmitBlockRequest) GetBlock() *Block {
	if x != nil {
		return x.Block
	}
	return nil
}

func (x *SubmitBlockRequest) GetBlockHash() []byte {
	if x != nil {
		return x.BlockHash
	}
	return nil
}

func (x *SubmitBlockRequest) GetNonce() int32 {
	if x != nil {
		return x.Nonce
	}
	return 0
}

// Response after submitting a block
type SubmitBlockResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	BlockHash     string                 `protobuf:"bytes,2,opt,name=block_hash,json=blockHash,proto3" json:"block_hash,omitempty"`
	ErrorMessage  string                 `protobuf:"bytes,3,opt,name=error_message,json=errorMessage,proto3" json:"error_message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubmitBlockResponse) Reset() {
	*x = SubmitBlockResponse{}
	mi := &file_proto_mining_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubmitBlockResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmitBlockResponse) ProtoMessage() {}

func (x *SubmitBlockResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_mining_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SubmitBlockResponse.ProtoReflect.Descriptor instead.
func (*SubmitBlockResponse) Descriptor() ([]byte, []int) {
	return file_proto_mining_proto_rawDescGZIP(), []int{4}
}

func (x *SubmitBlockResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *SubmitBlockResponse) GetBlockHash() string {
	if x != nil {
		return x.BlockHash
	}
	return ""
}

func (x *SubmitBlockResponse) GetErrorMessage() string {
	if x != nil {
		return x.ErrorMessage
	}
	return ""
}

// Transaction data
type Transaction struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	From          string                 `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	To            string                 `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty"`
	Amount        float32                `protobuf:"fixed32,3,opt,name=amount,proto3" json:"amount,omitempty"`
	Fee           float32                `protobuf:"fixed32,4,opt,name=fee,proto3" json:"fee,omitempty"` // Transaction fee
	Signature     []byte                 `protobuf:"bytes,5,opt,name=signature,proto3" json:"signature,omitempty"`
	TransactionId []byte                 `protobuf:"bytes,6,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	Vin           []*TXInput             `protobuf:"bytes,7,rep,name=vin,proto3" json:"vin,omitempty"`
	Vout          []*TXOutput            `protobuf:"bytes,8,rep,name=vout,proto3" json:"vout,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Transaction) Reset() {
	*x = Transaction{}
	mi := &file_proto_mining_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Transaction) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Transaction) ProtoMessage() {}

func (x *Transaction) ProtoReflect() protoreflect.Message {
	mi := &file_proto_mining_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Transaction.ProtoReflect.Descriptor instead.
func (*Transaction) Descriptor() ([]byte, []int) {
	return file_proto_mining_proto_rawDescGZIP(), []int{5}
}

func (x *Transaction) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *Transaction) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *Transaction) GetAmount() float32 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *Transaction) GetFee() float32 {
	if x != nil {
		return x.Fee
	}
	return 0
}

func (x *Transaction) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *Transaction) GetTransactionId() []byte {
	if x != nil {
		return x.TransactionId
	}
	return nil
}

func (x *Transaction) GetVin() []*TXInput {
	if x != nil {
		return x.Vin
	}
	return nil
}

func (x *Transaction) GetVout() []*TXOutput {
	if x != nil {
		return x.Vout
	}
	return nil
}

// Transaction Input
type TXInput struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Txid          []byte                 `protobuf:"bytes,1,opt,name=txid,proto3" json:"txid,omitempty"`
	Vout          int32                  `protobuf:"varint,2,opt,name=vout,proto3" json:"vout,omitempty"`
	Signature     []byte                 `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
	PubKey        []byte                 `protobuf:"bytes,4,opt,name=pub_key,json=pubKey,proto3" json:"pub_key,omitempty"` // Public key bytes
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TXInput) Reset() {
	*x = TXInput{}
	mi := &file_proto_mining_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TXInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TXInput) ProtoMessage() {}

func (x *TXInput) ProtoReflect() protoreflect.Message {
	mi := &file_proto_mining_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TXInput.ProtoReflect.Descriptor instead.
func (*TXInput) Descriptor() ([]byte, []int) {
	return file_proto_mining_proto_rawDescGZIP(), []int{6}
}

func (x *TXInput) GetTxid() []byte {
	if x != nil {
		return x.Txid
	}
	return nil
}

func (x *TXInput) GetVout() int32 {
	if x != nil {
		return x.Vout
	}
	return 0
}

func (x *TXInput) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

func (x *TXInput) GetPubKey() []byte {
	if x != nil {
		return x.PubKey
	}
	return nil
}

// Transaction Output
type TXOutput struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Value         float32                `protobuf:"fixed32,1,opt,name=value,proto3" json:"value,omitempty"`
	Address       string                 `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TXOutput) Reset() {
	*x = TXOutput{}
	mi := &file_proto_mining_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TXOutput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TXOutput) ProtoMessage() {}

func (x *TXOutput) ProtoReflect() protoreflect.Message {
	mi := &file_proto_mining_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TXOutput.ProtoReflect.Descriptor instead.
func (*TXOutput) Descriptor() ([]byte, []int) {
	return file_proto_mining_proto_rawDescGZIP(), []int{7}
}

func (x *TXOutput) GetValue() float32 {
	if x != nil {
		return x.Value
	}
	return 0
}

func (x *TXOutput) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

// Request to get blockchain status
type BlockchainStatusRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *BlockchainStatusRequest) Reset() {
	*x = BlockchainStatusRequest{}
	mi := &file_proto_mining_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BlockchainStatusRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockchainStatusRequest) ProtoMessage() {}

func (x *BlockchainStatusRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_mining_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockchainStatusRequest.ProtoReflect.Descriptor instead.
func (*BlockchainStatusRequest) Descriptor() ([]byte, []int) {
	return file_proto_mining_proto_rawDescGZIP(), []int{8}
}

// Response containing blockchain status
type BlockchainStatusResponse struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	Height          int32                  `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
	LatestBlockHash string                 `protobuf:"bytes,2,opt,name=latest_block_hash,json=latestBlockHash,proto3" json:"latest_block_hash,omitempty"`
	Difficulty      int32                  `protobuf:"varint,3,opt,name=difficulty,proto3" json:"difficulty,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *BlockchainStatusResponse) Reset() {
	*x = BlockchainStatusResponse{}
	mi := &file_proto_mining_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *BlockchainStatusResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BlockchainStatusResponse) ProtoMessage() {}

func (x *BlockchainStatusResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_mining_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BlockchainStatusResponse.ProtoReflect.Descriptor instead.
func (*BlockchainStatusResponse) Descriptor() ([]byte, []int) {
	return file_proto_mining_proto_rawDescGZIP(), []int{9}
}

func (x *BlockchainStatusResponse) GetHeight() int32 {
	if x != nil {
		return x.Height
	}
	return 0
}

func (x *BlockchainStatusResponse) GetLatestBlockHash() string {
	if x != nil {
		return x.LatestBlockHash
	}
	return ""
}

func (x *BlockchainStatusResponse) GetDifficulty() int32 {
	if x != nil {
		return x.Difficulty
	}
	return 0
}

// Request to get pending transactions
type PendingTransactionsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PendingTransactionsRequest) Reset() {
	*x = PendingTransactionsRequest{}
	mi := &file_proto_mining_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PendingTransactionsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PendingTransactionsRequest) ProtoMessage() {}

func (x *PendingTransactionsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_mining_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PendingTransactionsRequest.ProtoReflect.Descriptor instead.
func (*PendingTransactionsRequest) Descriptor() ([]byte, []int) {
	return file_proto_mining_proto_rawDescGZIP(), []int{10}
}

// Response containing pending transactions
type PendingTransactionsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Transactions  []*Transaction         `protobuf:"bytes,1,rep,name=transactions,proto3" json:"transactions,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PendingTransactionsResponse) Reset() {
	*x = PendingTransactionsResponse{}
	mi := &file_proto_mining_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PendingTransactionsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PendingTransactionsResponse) ProtoMessage() {}

func (x *PendingTransactionsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_mining_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PendingTransactionsResponse.ProtoReflect.Descriptor instead.
func (*PendingTransactionsResponse) Descriptor() ([]byte, []int) {
	return file_proto_mining_proto_rawDescGZIP(), []int{11}
}

func (x *PendingTransactionsResponse) GetTransactions() []*Transaction {
	if x != nil {
		return x.Transactions
	}
	return nil
}

var File_proto_mining_proto protoreflect.FileDescriptor

const file_proto_mining_proto_rawDesc = "" +
	"\n" +
	"\x12proto/mining.proto\x12\x05proto\";\n" +
	"\x14BlockTemplateRequest\x12#\n" +
	"\rminer_address\x18\x01 \x01(\tR\fminerAddress\"\x9d\x01\n" +
	"\x05Block\x12\x1c\n" +
	"\ttimestamp\x18\x01 \x01(\x03R\ttimestamp\x12&\n" +
	"\x0fprev_block_hash\x18\x02 \x01(\fR\rprevBlockHash\x12\x16\n" +
	"\x06height\x18\x03 \x01(\x05R\x06height\x126\n" +
	"\ftransactions\x18\x04 \x03(\v2\x12.proto.TransactionR\ftransactions\"[\n" +
	"\x15BlockTemplateResponse\x12\"\n" +
	"\x05block\x18\x01 \x01(\v2\f.proto.BlockR\x05block\x12\x1e\n" +
	"\n" +
	"difficulty\x18\x02 \x01(\x05R\n" +
	"difficulty\"m\n" +
	"\x12SubmitBlockRequest\x12\"\n" +
	"\x05block\x18\x01 \x01(\v2\f.proto.BlockR\x05block\x12\x1d\n" +
	"\n" +
	"block_hash\x18\x02 \x01(\fR\tblockHash\x12\x14\n" +
	"\x05nonce\x18\x03 \x01(\x05R\x05nonce\"s\n" +
	"\x13SubmitBlockResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\x12\x1d\n" +
	"\n" +
	"block_hash\x18\x02 \x01(\tR\tblockHash\x12#\n" +
	"\rerror_message\x18\x03 \x01(\tR\ferrorMessage\"\xe7\x01\n" +
	"\vTransaction\x12\x12\n" +
	"\x04from\x18\x01 \x01(\tR\x04from\x12\x0e\n" +
	"\x02to\x18\x02 \x01(\tR\x02to\x12\x16\n" +
	"\x06amount\x18\x03 \x01(\x02R\x06amount\x12\x10\n" +
	"\x03fee\x18\x04 \x01(\x02R\x03fee\x12\x1c\n" +
	"\tsignature\x18\x05 \x01(\fR\tsignature\x12%\n" +
	"\x0etransaction_id\x18\x06 \x01(\fR\rtransactionId\x12 \n" +
	"\x03vin\x18\a \x03(\v2\x0e.proto.TXInputR\x03vin\x12#\n" +
	"\x04vout\x18\b \x03(\v2\x0f.proto.TXOutputR\x04vout\"h\n" +
	"\aTXInput\x12\x12\n" +
	"\x04txid\x18\x01 \x01(\fR\x04txid\x12\x12\n" +
	"\x04vout\x18\x02 \x01(\x05R\x04vout\x12\x1c\n" +
	"\tsignature\x18\x03 \x01(\fR\tsignature\x12\x17\n" +
	"\apub_key\x18\x04 \x01(\fR\x06pubKey\":\n" +
	"\bTXOutput\x12\x14\n" +
	"\x05value\x18\x01 \x01(\x02R\x05value\x12\x18\n" +
	"\aaddress\x18\x02 \x01(\tR\aaddress\"\x19\n" +
	"\x17BlockchainStatusRequest\"~\n" +
	"\x18BlockchainStatusResponse\x12\x16\n" +
	"\x06height\x18\x01 \x01(\x05R\x06height\x12*\n" +
	"\x11latest_block_hash\x18\x02 \x01(\tR\x0flatestBlockHash\x12\x1e\n" +
	"\n" +
	"difficulty\x18\x03 \x01(\x05R\n" +
	"difficulty\"\x1c\n" +
	"\x1aPendingTransactionsRequest\"U\n" +
	"\x1bPendingTransactionsResponse\x126\n" +
	"\ftransactions\x18\x01 \x03(\v2\x12.proto.TransactionR\ftransactions2\xe5\x02\n" +
	"\rMiningService\x12O\n" +
	"\x10GetBlockTemplate\x12\x1b.proto.BlockTemplateRequest\x1a\x1c.proto.BlockTemplateResponse\"\x00\x12F\n" +
	"\vSubmitBlock\x12\x19.proto.SubmitBlockRequest\x1a\x1a.proto.SubmitBlockResponse\"\x00\x12X\n" +
	"\x13GetBlockchainStatus\x12\x1e.proto.BlockchainStatusRequest\x1a\x1f.proto.BlockchainStatusResponse\"\x00\x12a\n" +
	"\x16GetPendingTransactions\x12!.proto.PendingTransactionsRequest\x1a\".proto.PendingTransactionsResponse\"\x00B\tZ\a./protob\x06proto3"

var (
	file_proto_mining_proto_rawDescOnce sync.Once
	file_proto_mining_proto_rawDescData []byte
)

func file_proto_mining_proto_rawDescGZIP() []byte {
	file_proto_mining_proto_rawDescOnce.Do(func() {
		file_proto_mining_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_proto_mining_proto_rawDesc), len(file_proto_mining_proto_rawDesc)))
	})
	return file_proto_mining_proto_rawDescData
}

var file_proto_mining_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_proto_mining_proto_goTypes = []any{
	(*BlockTemplateRequest)(nil),        // 0: proto.BlockTemplateRequest
	(*Block)(nil),                       // 1: proto.Block
	(*BlockTemplateResponse)(nil),       // 2: proto.BlockTemplateResponse
	(*SubmitBlockRequest)(nil),          // 3: proto.SubmitBlockRequest
	(*SubmitBlockResponse)(nil),         // 4: proto.SubmitBlockResponse
	(*Transaction)(nil),                 // 5: proto.Transaction
	(*TXInput)(nil),                     // 6: proto.TXInput
	(*TXOutput)(nil),                    // 7: proto.TXOutput
	(*BlockchainStatusRequest)(nil),     // 8: proto.BlockchainStatusRequest
	(*BlockchainStatusResponse)(nil),    // 9: proto.BlockchainStatusResponse
	(*PendingTransactionsRequest)(nil),  // 10: proto.PendingTransactionsRequest
	(*PendingTransactionsResponse)(nil), // 11: proto.PendingTransactionsResponse
}
var file_proto_mining_proto_depIdxs = []int32{
	5,  // 0: proto.Block.transactions:type_name -> proto.Transaction
	1,  // 1: proto.BlockTemplateResponse.block:type_name -> proto.Block
	1,  // 2: proto.SubmitBlockRequest.block:type_name -> proto.Block
	6,  // 3: proto.Transaction.vin:type_name -> proto.TXInput
	7,  // 4: proto.Transaction.vout:type_name -> proto.TXOutput
	5,  // 5: proto.PendingTransactionsResponse.transactions:type_name -> proto.Transaction
	0,  // 6: proto.MiningService.GetBlockTemplate:input_type -> proto.BlockTemplateRequest
	3,  // 7: proto.MiningService.SubmitBlock:input_type -> proto.SubmitBlockRequest
	8,  // 8: proto.MiningService.GetBlockchainStatus:input_type -> proto.BlockchainStatusRequest
	10, // 9: proto.MiningService.GetPendingTransactions:input_type -> proto.PendingTransactionsRequest
	2,  // 10: proto.MiningService.GetBlockTemplate:output_type -> proto.BlockTemplateResponse
	4,  // 11: proto.MiningService.SubmitBlock:output_type -> proto.SubmitBlockResponse
	9,  // 12: proto.MiningService.GetBlockchainStatus:output_type -> proto.BlockchainStatusResponse
	11, // 13: proto.MiningService.GetPendingTransactions:output_type -> proto.PendingTransactionsResponse
	10, // [10:14] is the sub-list for method output_type
	6,  // [6:10] is the sub-list for method input_type
	6,  // [6:6] is the sub-list for extension type_name
	6,  // [6:6] is the sub-list for extension extendee
	0,  // [0:6] is the sub-list for field type_name
}

func init() { file_proto_mining_proto_init() }
func file_proto_mining_proto_init() {
	if File_proto_mining_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_proto_mining_proto_rawDesc), len(file_proto_mining_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_mining_proto_goTypes,
		DependencyIndexes: file_proto_mining_proto_depIdxs,
		MessageInfos:      file_proto_mining_proto_msgTypes,
	}.Build()
	File_proto_mining_proto = out.File
	file_proto_mining_proto_goTypes = nil
	file_proto_mining_proto_depIdxs = nil
}
