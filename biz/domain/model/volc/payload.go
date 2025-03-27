package volc

type TTSRequest struct {
	User *TTSUser `protobuf:"bytes,1,opt,name=user,proto3" json:"user,omitempty"`
	// Similar to TTSResponse.event field.
	Event     int32         `protobuf:"varint,2,opt,name=event,proto3" json:"event,omitempty"`
	Namespace string        `protobuf:"bytes,3,opt,name=namespace,proto3" json:"namespace,omitempty"`
	ReqParams *TTSReqParams `protobuf:"bytes,10,opt,name=req_params,json=reqParams,proto3" json:"req_params,omitempty"`
}

type TTSUser struct {
	Uid            string `protobuf:"bytes,1,opt,name=uid,proto3" json:"uid,omitempty"`
	Did            string `protobuf:"bytes,2,opt,name=did,proto3" json:"did,omitempty"`
	DevicePlatform string `protobuf:"bytes,3,opt,name=device_platform,json=devicePlatform,proto3" json:"device_platform,omitempty"`
	DeviceType     string `protobuf:"bytes,4,opt,name=device_type,json=deviceType,proto3" json:"device_type,omitempty"`
	VersionCode    string `protobuf:"bytes,5,opt,name=version_code,json=versionCode,proto3" json:"version_code,omitempty"`
	Language       string `protobuf:"bytes,6,opt,name=language,proto3" json:"language,omitempty"`
}

type TTSReqParams struct {
	Text           string            `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
	Texts          *Texts            `protobuf:"bytes,2,opt,name=texts,proto3" json:"texts,omitempty"`
	Ssml           string            `protobuf:"bytes,3,opt,name=ssml,proto3" json:"ssml,omitempty"`
	Speaker        string            `protobuf:"bytes,4,opt,name=speaker,proto3" json:"speaker,omitempty"`
	AudioParams    *AudioParams      `protobuf:"bytes,5,opt,name=audio_params,json=audioParams,proto3" json:"audio_params,omitempty"`
	EngineParams   *EngineParams     `protobuf:"bytes,6,opt,name=engine_params,json=engineParams,proto3" json:"engine_params,omitempty"`
	EnableAudio2Bs bool              `protobuf:"varint,7,opt,name=enable_audio2bs,json=enableAudio2bs,proto3" json:"enable_audio2bs,omitempty"`
	EnableTextSeg  bool              `protobuf:"varint,8,opt,name=enable_text_seg,json=enableTextSeg,proto3" json:"enable_text_seg,omitempty"`
	Additions      map[string]string `protobuf:"bytes,100,rep,name=additions,proto3" json:"additions,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

type Texts struct {
	Texts []string `protobuf:"bytes,1,rep,name=texts,proto3" json:"texts,omitempty"`
}

type AudioParams struct {
	Format          string `protobuf:"bytes,1,opt,name=format,proto3" json:"format,omitempty"`
	SampleRate      int32  `protobuf:"varint,2,opt,name=sample_rate,json=sampleRate,proto3" json:"sample_rate,omitempty"`
	Channel         int32  `protobuf:"varint,3,opt,name=channel,proto3" json:"channel,omitempty"`
	SpeechRate      int32  `protobuf:"varint,4,opt,name=speech_rate,json=speechRate,proto3" json:"speech_rate,omitempty"`
	PitchRate       int32  `protobuf:"varint,5,opt,name=pitch_rate,json=pitchRate,proto3" json:"pitch_rate,omitempty"`
	BitRate         int32  `protobuf:"varint,6,opt,name=bit_rate,json=bitRate,proto3" json:"bit_rate,omitempty"`
	Volume          int32  `protobuf:"varint,7,opt,name=volume,proto3" json:"volume,omitempty"`
	Lang            string `protobuf:"bytes,8,opt,name=lang,proto3" json:"lang,omitempty"`
	Emotion         string `protobuf:"bytes,9,opt,name=emotion,proto3" json:"emotion,omitempty"`
	Gender          string `protobuf:"bytes,10,opt,name=gender,proto3" json:"gender,omitempty"`
	EnableTimestamp bool   `protobuf:"varint,11,opt,name=enable_timestamp,json=enableTimestamp,proto3" json:"enable_timestamp,omitempty"`
}

type EngineParams struct {
	EngineContext                string   `protobuf:"bytes,1,opt,name=engine_context,json=engineContext,proto3" json:"engine_context,omitempty"`
	PhonemeSize                  string   `protobuf:"bytes,2,opt,name=phoneme_size,json=phonemeSize,proto3" json:"phoneme_size,omitempty"`
	EnableFastTextSeg            bool     `protobuf:"varint,3,opt,name=enable_fast_text_seg,json=enableFastTextSeg,proto3" json:"enable_fast_text_seg,omitempty"`
	ForceBreak                   bool     `protobuf:"varint,4,opt,name=force_break,json=forceBreak,proto3" json:"force_break,omitempty"`
	BreakByProsody               int32    `protobuf:"varint,5,opt,name=break_by_prosody,json=breakByProsody,proto3" json:"break_by_prosody,omitempty"`
	EnableEngineDebugInfo        bool     `protobuf:"varint,6,opt,name=enable_engine_debug_info,json=enableEngineDebugInfo,proto3" json:"enable_engine_debug_info,omitempty"`
	FlushSentence                bool     `protobuf:"varint,7,opt,name=flush_sentence,json=flushSentence,proto3" json:"flush_sentence,omitempty"`
	LabVersion                   string   `protobuf:"bytes,8,opt,name=lab_version,json=labVersion,proto3" json:"lab_version,omitempty"`
	EnableIpaExtraction          bool     `protobuf:"varint,9,opt,name=enable_ipa_extraction,json=enableIpaExtraction,proto3" json:"enable_ipa_extraction,omitempty"`
	EnableNaiveTn                bool     `protobuf:"varint,10,opt,name=enable_naive_tn,json=enableNaiveTn,proto3" json:"enable_naive_tn,omitempty"`
	EnableLatexTn                bool     `protobuf:"varint,11,opt,name=enable_latex_tn,json=enableLatexTn,proto3" json:"enable_latex_tn,omitempty"`
	DisableNewlineStrategy       bool     `protobuf:"varint,12,opt,name=disable_newline_strategy,json=disableNewlineStrategy,proto3" json:"disable_newline_strategy,omitempty"`
	SupportedLanguages           []string `protobuf:"bytes,13,rep,name=supported_languages,json=supportedLanguages,proto3" json:"supported_languages,omitempty"`
	ContextLanguage              string   `protobuf:"bytes,14,opt,name=context_language,json=contextLanguage,proto3" json:"context_language,omitempty"`
	ContextTexts                 []string `protobuf:"bytes,15,rep,name=context_texts,json=contextTexts,proto3" json:"context_texts,omitempty"`
	EnableRecoverPuncts          bool     `protobuf:"varint,16,opt,name=enable_recover_puncts,json=enableRecoverPuncts,proto3" json:"enable_recover_puncts,omitempty"`
	EosProsody                   int32    `protobuf:"varint,17,opt,name=eos_prosody,json=eosProsody,proto3" json:"eos_prosody,omitempty"`
	PrependSilenceSeconds        float64  `protobuf:"fixed64,18,opt,name=prepend_silence_seconds,json=prependSilenceSeconds,proto3" json:"prepend_silence_seconds,omitempty"`
	MaxParagraphPhonemeSize      int32    `protobuf:"varint,19,opt,name=max_paragraph_phoneme_size,json=maxParagraphPhonemeSize,proto3" json:"max_paragraph_phoneme_size,omitempty"`
	ParagraphSubSentences        []string `protobuf:"bytes,20,rep,name=paragraph_sub_sentences,json=paragraphSubSentences,proto3" json:"paragraph_sub_sentences,omitempty"`
	MaxLengthToFilterParenthesis int32    `protobuf:"varint,21,opt,name=max_length_to_filter_parenthesis,json=maxLengthToFilterParenthesis,proto3" json:"max_length_to_filter_parenthesis,omitempty"`
	EnableLanguageDetector       bool     `protobuf:"varint,22,opt,name=enable_language_detector,json=enableLanguageDetector,proto3" json:"enable_language_detector,omitempty"`
}
