/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"encoding/json"
	"io"
	"log"
	"net"

	pb "example.com/grpc-fabric-test/helloworld"

	"google.golang.org/grpc"
)

const (
	port = ":50053"
)

type HelloStruct struct {
	Language string
	Hello    string
}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer

	greetingMap map[string]string
}

var workerNum uint32 = 1

func (s *server) loadWords() {
	log.Print("Loading words\n")
	data := helloWords
	helloStruct := []HelloStruct{}

	if err := json.Unmarshal(data, &helloStruct); err != nil {
		log.Fatalf("Failed to load default words: %v", err)
	}

	for _, greeting := range helloStruct {
		s.greetingMap[greeting.Language] = greeting.Hello
	}
}

func newServer() *server {
	log.Print("Create new server\n")
	s := &server{greetingMap: make(map[string]string)}
	s.loadWords()
	return s
}

/* LanguageService

 */
func (s *server) LanguageService(stream pb.Greeter_LanguageServiceServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Print("EOF")

			continue
		}

		if err != nil {
			log.Print(err)
			continue
		}

		greeting := s.greetingMap[in.Language]
		log.Print("Request greeting in ", in.GetLanguage(), ": ", greeting)
		if err := stream.Send(&pb.LanguageReply{Id: in.Id, Greeting: greeting}); err != nil {
			log.Print(err)

			continue
		}
	}
}

func main() {
	lis, err := net.Listen("tcp4", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)

	pb.RegisterGreeterServer(s, newServer())
	log.Print("Ready to serve!\n===============================\n")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

var helloWords = []byte(`
[{
    "language": "English",
    "hello": "Welcome!"
  },
  {
    "language": "Afrikaans",
    "hello": "hallo"
  },
  {
    "language": "Albanian",
    "hello": "Përshëndetje"
  },
  {
    "language": "Amharic",
    "hello": "ሰላም"
  },
  {
    "language": "Arabic",
    "hello": "مرحبا"
  },
  {
    "language": "Armenian",
    "hello": "Բարեւ"
  },
  {
    "language": "Azerbaijani",
    "hello": "Salam"
  },
  {
    "language": "Basque",
    "hello": "Kaixo"
  },
  {
    "language": "Belarusian",
    "hello": "добры дзень"
  },
  {
    "language": "Bengali",
    "hello": "হ্যালো"
  },
  {
    "language": "Bosnian",
    "hello": "zdravo"
  },
  {
    "language": "Bulgarian",
    "hello": "Здравейте"
  },
  {
    "language": "Catalan",
    "hello": "Hola"
  },
  {
    "language": "Cebuano",
    "hello": "Hello"
  },
  {
    "language": "Chichewa",
    "hello": "Moni"
  },
  {
    "language": "Chinese (Simplified)",
    "hello": "您好"
  },
  {
    "language": "Chinese (Traditional)",
    "hello": "您好"
  },
  {
    "language": "Corsican",
    "hello": "Bonghjornu"
  },
  {
    "language": "Croatian",
    "hello": "zdravo"
  },
  {
    "language": "Czech",
    "hello": "Ahoj"
  },
  {
    "language": "Danish",
    "hello": "Hej"
  },
  {
    "language": "Dutch",
    "hello": "Hallo"
  },
  {
    "language": "English",
    "hello": "Hello"
  },
  {
    "language": "Esperanto",
    "hello": "Saluton"
  },
  {
    "language": "Estonian",
    "hello": "Tere"
  },
  {
    "language": "Filipino",
    "hello": "Hello"
  },
  {
    "language": "Finnish",
    "hello": "Hei"
  },
  {
    "language": "French",
    "hello": "Bonjour"
  },
  {
    "language": "Frisian",
    "hello": "Hello"
  },
  {
    "language": "Galician",
    "hello": "Ola"
  },
  {
    "language": "Georgian",
    "hello": "გამარჯობა"
  },
  {
    "language": "German",
    "hello": "Hallo"
  },
  {
    "language": "Greek",
    "hello": "Γεια σας"
  },
  {
    "language": "Gujarati",
    "hello": "હેલો"
  },
  {
    "language": "Haitian Creole",
    "hello": "Bonjou"
  },
  {
    "language": "Hausa",
    "hello": "Sannu"
  },
  {
    "language": "Hawaiian",
    "hello": "Alohaʻoe"
  },
  {
    "language": "Hebrew",
    "hello": "שלום"
  },
  {
    "language": "Hindi",
    "hello": "नमस्ते"
  },
  {
    "language": "Hmong",
    "hello": "Nyob zoo"
  },
  {
    "language": "Hungarian",
    "hello": "Helló"
  },
  {
    "language": "Icelandic",
    "hello": "Halló"
  },
  {
    "language": "Igbo",
    "hello": "Ndewo"
  },
  {
    "language": "Indonesian",
    "hello": "Halo"
  },
  {
    "language": "Irish",
    "hello": "Dia duit"
  },
  {
    "language": "Italian",
    "hello": "Ciao"
  },
  {
    "language": "Japanese",
    "hello": "こんにちは"
  },
  {
    "language": "Javanese",
    "hello": "Hello"
  },
  {
    "language": "Kannada",
    "hello": "ಹಲೋ"
  },
  {
    "language": "Kazakh",
    "hello": "Сәлем"
  },
  {
    "language": "Khmer",
    "hello": "ជំរាបសួរ"
  },
  {
    "language": "Korean",
    "hello": "안녕하세요."
  },
  {
    "language": "Kurdish (Kurmanji)",
    "hello": "Hello"
  },
  {
    "language": "Kyrgyz",
    "hello": "салам"
  },
  {
    "language": "Lao",
    "hello": "ສະບາຍດີ"
  },
  {
    "language": "Latin",
    "hello": "salve"
  },
  {
    "language": "Latvian",
    "hello": "Labdien!"
  },
  {
    "language": "Lithuanian",
    "hello": "Sveiki"
  },
  {
    "language": "Luxembourgish",
    "hello": "Moien"
  },
  {
    "language": "Macedonian",
    "hello": "Здраво"
  },
  {
    "language": "Malagasy",
    "hello": "Hello"
  },
  {
    "language": "Malay",
    "hello": "Hello"
  },
  {
    "language": "Malayalam",
    "hello": "ഹലോ"
  },
  {
    "language": "Maltese",
    "hello": "Hello"
  },
  {
    "language": "Maori",
    "hello": "Hiha"
  },
  {
    "language": "Marathi",
    "hello": "हॅलो"
  },
  {
    "language": "Mongolian",
    "hello": "Сайн байна уу"
  },
  {
    "language": "Myanmar (Burmese)",
    "hello": "မင်္ဂလာပါ"
  },
  {
    "language": "Nepali",
    "hello": "नमस्ते"
  },
  {
    "language": "Norwegian",
    "hello": "Hallo"
  },
  {
    "language": "Pashto",
    "hello": "سلام"
  },
  {
    "language": "Persian",
    "hello": "سلام"
  },
  {
    "language": "Polish",
    "hello": "Cześć"
  },
  {
    "language": "Portuguese",
    "hello": "Olá"
  },
  {
    "language": "Punjabi",
    "hello": "ਹੈਲੋ"
  },
  {
    "language": "Romanian",
    "hello": "Alo"
  },
  {
    "language": "Russian",
    "hello": "привет"
  },
  {
    "language": "Samoan",
    "hello": "Talofa"
  },
  {
    "language": "Scots Gaelic",
    "hello": "Hello"
  },
  {
    "language": "Serbian",
    "hello": "Здраво"
  },
  {
    "language": "Sesotho",
    "hello": "Hello"
  },
  {
    "language": "Shona",
    "hello": "Hello"
  },
  {
    "language": "Sindhi",
    "hello": "هيلو"
  },
  {
    "language": "Sinhala",
    "hello": "හෙලෝ"
  },
  {
    "language": "Slovak",
    "hello": "ahoj"
  },
  {
    "language": "Slovenian",
    "hello": "Pozdravljeni"
  },
  {
    "language": "Somali",
    "hello": "Hello"
  },
  {
    "language": "Spanish",
    "hello": "Hola"
  },
  {
    "language": "Sundanese",
    "hello": "halo"
  },
  {
    "language": "Swahili",
    "hello": "Sawa"
  },
  {
    "language": "Swedish",
    "hello": "Hallå"
  },
  {
    "language": "Tajik",
    "hello": "Салом"
  },
  {
    "language": "Tamil",
    "hello": "ஹலோ"
  },
  {
    "language": "Telugu",
    "hello": "హలో"
  },
  {
    "language": "Thai",
    "hello": "สวัสดี"
  },
  {
    "language": "Turkish",
    "hello": "Merhaba"
  },
  {
    "language": "Ukranian",
    "hello": "Здрастуйте"
  },
  {
    "language": "Urdu",
    "hello": "ہیلو"
  },
  {
    "language": "Uzbek",
    "hello": "Salom"
  },
  {
    "language": "Vietnamese",
    "hello": "Xin chào"
  },
  {
    "language": "Welsh",
    "hello": "Helo"
  },
  {
    "language": "Xhosa",
    "hello": "Sawubona"
  },
  {
    "language": "Yiddish",
    "hello": "העלא"
  },
  {
    "language": "Yoruba",
    "hello": "Kaabo"
  },
  {
    "language": "Zulu",
    "hello": "Sawubona"
  }
]
`)
