package main

type Table string
type Column string

type ForeignKeys map[Column]Table

type Config map[Table]ForeignKeys
