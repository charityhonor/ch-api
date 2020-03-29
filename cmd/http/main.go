package main

func main() {
	s := &Server{}
	if err := s.ParseFlags(); err != nil {
		panic(err)
	}

	s.AddRoutes(
		DriveRoutes,
	)

	if err := s.Run(); err != nil {
		panic(err)
	}
}
