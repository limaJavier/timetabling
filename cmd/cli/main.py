from enum import Enum
import json
import os
import subprocess
import tempfile
from reportlab.lib.pagesizes import A4
from reportlab.platypus import SimpleDocTemplate, Table, TableStyle, Spacer
from reportlab.lib import colors


class ExportType(Enum):
    PDF = "pdf"
    JSON = "json"


class ResultType(Enum):
    SATISFIABLE = 10
    UNSATISFIABLE = 20
    VERIFICATION_FAILED = 15
    ERROR = 1


class Strategy(Enum):
    PURE = "pure"
    POSTPONED = "postponed"
    HYBRID = "hybrid"


class SolverType(Enum):
    KISSAT = "kissat"
    CADICAL = "cadical"
    MINISAT = "minisat"
    CRYPTOMINISAT = "cryptominisat"
    GLUCOSE_SIMP = "glucosesimp"
    GLUCOSE_SYRUP = "glucosesyrup"
    SLIME = "slime"
    ORTOOLSAT = "ortoolsat"


_DAYS = {
    0: "Lunes",
    1: "Martes",
    2: "Miércoles",
    3: "Jueves",
    4: "Viernes",
    5: "Sábado",
    6: "Domingo",
}


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
    def __init__(self, timetabler_path: str) -> None:
        self._timetable: dict[str, list[dict[str, int]]]
        self._timetabler_path = timetabler_path

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

    def build_timetable(
        self,
        strategy: str = Strategy.PURE.value,
        solver: str = SolverType.CADICAL.value,
        similarity: str = "0.5",
    ) -> tuple[int, str]:
        # Prepare input as json
        cli_input = json.dumps(
            {
                "subjects": [subject.__dict__ for subject in self._subjects],
                "professors": [professor.__dict__ for professor in self._professors],
                "rooms": [room.__dict__ for room in self._rooms],
                "classes": [class_.__dict__ for class_ in self._classes],
                "entries": [entry.__dict__ for entry in self._entries],
            }
        )

        # Create temporary file for timetabler's input
        with tempfile.NamedTemporaryFile(
            delete=False, mode="w", suffix=".json"
        ) as temp_file:
            temp_file.write(cli_input)
            temp_file_path = temp_file.name

        try:
            arguments = [
                f"{self._timetabler_path}",
                "-file",
                f"{temp_file_path}",
                "-strategy",
                f"{strategy}",
                "-solver",
                f"{solver}",
                "-similarity",
                f"{similarity}",
            ]
            process = subprocess.Popen(
                arguments,
                stdin=subprocess.PIPE,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                text=True,
            )

            stdout, _ = process.communicate()

            if process.returncode == ResultType.ERROR.value:
                return ResultType.ERROR.value, stdout
            elif process.returncode == ResultType.VERIFICATION_FAILED.value:
                return ResultType.VERIFICATION_FAILED.value, ""
            elif process.returncode == ResultType.UNSATISFIABLE.value:
                return ResultType.UNSATISFIABLE.value, ""

            raw_timetable: dict[str, list] = json.loads(stdout.split("\n")[0])
            self._timetable = {}

            for key, value in raw_timetable.items():
                key = self._classes[int(key)].name
                self._timetable[key] = []
                for lesson in value:
                    self._timetable[key].append(
                        {
                            "period": lesson["period"],
                            "day": lesson["day"],
                            "subject": self._subjects[lesson["subject"]].name,
                            "professor": self._professors[lesson["professor"]].name,
                            "room": self._rooms[lesson["room"]].name,
                        }
                    )

            return ResultType.SATISFIABLE.value, ""
        finally:
            os.remove(
                temp_file_path
            )  # Ensure temporary file is removed whatever happens

    def load(self, filename: str) -> None:
        """
        Load model from json file
        """
        # Reset entities
        self._subjects = []
        self._professors = []
        self._rooms = []
        self._classes = []
        self._entries = []

        # Load json
        with open(filename, "r") as file:
            data = json.load(file)

        # Fill entities
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

    def export(self, file_type: str, filename: str):
        if file_type == ExportType.PDF.value:
            self._export_pdf(filename)
        elif file_type == ExportType.JSON.value:
            self._export_json(filename)
        else:
            raise Exception(f'file-type "{file_type}" is not expected')

    def _export_json(self, filename: str):
        with open(filename, "w") as file:
            json.dump(self._timetable, file)

    def _export_pdf(self, filename: str):
        # Create the PDF
        doc = SimpleDocTemplate(filename=filename, pagesize=A4)

        # Define styling
        style = TableStyle(
            [
                ("BACKGROUND", (0, 0), (-1, 0), colors.deepskyblue),
                ("TEXTCOLOR", (0, 0), (-1, 0), colors.whitesmoke),
                ("ALIGN", (0, 0), (-1, -1), "CENTER"),
                ("FONTNAME", (0, 0), (-1, 0), "Helvetica-Bold"),
                ("GRID", (0, 0), (-1, -1), 1, colors.black),
                ("FONTSIZE", (0, 0), (-1, -1), 7),
            ]
        )

        groups = list(self._timetable.keys())
        groups.sort()  # Sort groups by alphabetical order
        tables = []
        # Set column widths
        for group in groups:
            lessons = self._timetable[group]

            data = [
                ["" for _ in range(6)] for _ in range(7)
            ]  # Assuming there are 6 periods and 5 days
            data[0][0] = group.upper()

            for i in range(1, 6):
                data[0][i] = _DAYS[i - 1]
            for i in range(1, 7):
                data[i][0] = f"{i}"

            for lesson in lessons:
                period = lesson["period"] + 1
                day = lesson["day"] + 1

                subject = lesson["subject"]
                professor = lesson["professor"]
                room = lesson["room"]

                data[period][day] = f"{subject}\n{professor}/{room}"

            table = Table(
                data, colWidths=[20] + [100] * 5
            )  # Create table with a fixed column-width
            table.setStyle(style)
            table.keepWithNext = True  # Prevent table from being split across pages
            tables.append(table)  # Add the table to the elements list
            tables.append(Spacer(1, 20))  # Add some space between tables

        doc.build(tables)  # Add the table to the document and build it

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


if __name__ == "__main__":
    model = Model("bin/timetabler")
    model.load("test/out/satisfiable/11_i.json")
    result, message = model.build_timetable(
        strategy=Strategy.HYBRID.value,
        solver=SolverType.CADICAL.value,
        similarity="0.75",
    )
    if result == ResultType.SATISFIABLE.value:
        model.export("json", "timetable.json")
        model.export("pdf", "timetable.pdf")
