import json


class Subject:
    def __init__(self, name: str) -> None:
        self.id: int | None = None
        self.name = name


class Professor:
    def __init__(self, name: str, availability: list[list[bool]]) -> None:
        self.id: int | None = None
        self.name = name
        self.availability = availability


class Room:
    def __init__(self, name: str, capacity: int) -> None:
        self.id: int | None = None
        self.name = name
        self.capacity = capacity


class Class:
    def __init__(self, name: str, size: int) -> None:
        self.id: int | None = None
        self.name = name
        self.size = size


class Entry:
    def __init__(
        self,
        subject: str,
        professor: str,
        classes: list[str],
        lessons: int,
        permissibility: list[list[bool]],
        rooms: list[str],
    ) -> None:
        self.subject: str | int = subject
        self.professor: str | int = professor
        self.classes: list[str] | list[int] = classes
        self.lessons = lessons
        self.permissibility = permissibility
        self.rooms: list[str] | list[int] = rooms


class Model:
    def __init__(self) -> None:
        self._subjects: list[Subject] = []
        self._professors: list[Professor] = []
        self._rooms: list[Room] = []
        self._classes: list[Class] = []
        self._entries: list[Entry] = []

    def add_subject(self, subject: Subject) -> None:
        self._add_entity(subject, self._subjects)

    def add_professor(self, professor: Professor) -> None:
        self._add_entity(professor, self._professors)

    def add_room(self, room: Room) -> None:
        self._add_entity(room, self._rooms)

    def add_class(self, class_: Class) -> None:
        self._add_entity(class_, self._classes)

    def add_entry(self, entry: Entry) -> None:
        entry.subject = self._get_index(str(entry.subject), self._subjects)
        entry.professor = self._get_index(str(entry.professor), self._professors)
        entry.classes = list(
            map(
                lambda class_name: self._get_index(str(class_name), self._classes),
                entry.classes,
            )
        )
        entry.rooms = list(
            map(
                lambda room_name: self._get_index(str(room_name), self._rooms),
                entry.rooms,
            )
        )

        # TODO: Verify that can only be one entry for each subject-professor and group
        self._entries.append(entry)

    def timetable(self) -> None:
        cli_input = json.dumps(
            {
                "subjects": [subject.__dict__ for subject in self._subjects],
                "professors": [professor.__dict__ for professor in self._professors],
                "rooms": [room.__dict__ for room in self._rooms],
                "classes": [class_.__dict__ for class_ in self._classes],
                "entries": [entry.__dict__ for entry in self._entries],
            }
        )
        print(cli_input)

    def load(self, filename: str) -> None:
        with open(filename, "r") as file:
            data = json.load(file)

        for subject in data["subjects"]:
            self.add_subject(Subject(subject["name"]))

        for professor in data["professors"]:
            self.add_professor(Professor(professor["name"], professor["availability"]))

        for room in data["rooms"]:
            self.add_room(Room(room["name"], room["capacity"]))

        for class_ in data["classes"]:
            self.add_class(Class(class_["name"], class_["size"]))

        for entry in data["entries"]:
            self.add_entry(
                Entry(
                    self._subjects[entry["subject"]].name,
                    self._professors[entry["professor"]].name,
                    list(
                        map(
                            lambda class_index: self._classes[class_index].name,
                            entry["classes"],
                        )
                    ),
                    entry["lessons"],
                    entry["permissibility"],
                    list(
                        map(
                            lambda room_index: self._rooms[room_index].name,
                            entry["rooms"],
                        )
                    ),
                )
            )

    def _add_entity(self, entity, registered_entities: list) -> None:
        if any(
            map(
                lambda registered_entity: registered_entity.name == entity.name,
                registered_entities,
            )
        ):
            raise Exception(f'{type(entity).__name__} "{entity.name}" already exists.')

        entity.id = len(registered_entities)
        registered_entities.append(entity)

    def _get_index(self, entity_name: str, entities: list) -> int:
        return list(map(lambda entity: entity.name, entities)).index(entity_name)
