# gokart

A golang library allowing to exctract meta data from GoPro video.

**WARNING**
This is a **Work In Progress**, code should be clean-up, documented, tested to be production ready.

## Tests

Some tests are available, create a `data` folder and download video example:
[20240914T1112_Ancenis.mp4](https://drive.google.com/file/d/15jsj2EUC1Xuy-kCqMmxUHkDCcQakrYSN/view?usp=drive_link)

```bash
mv ~/Downloads/20240914T1112_Ancenis.mp4 data
go test
```

## Examples

### DrawLap

An example is available to draw best lap trajectory from a video on an aerial image.
You will need the aerial image corresponding to the track in the video, for example:
[Ancenis.png](https://drive.google.com/file/d/1HnRUj4Lz5NOsJOMnHdNT5mKiQkN_wZnE/view?usp=drive_link)

```bash
mv ~/Downloads/Ancenis.png data
cd cmd/drawlap
go build
./drawlap -in ../../data/20240914T1112_Ancenis.mp4
```

It will generate an image named `best_lap.png` with trajectory in color where <span style="color:red">red</span> means deceleration and <span style="color:green">green</span> means acceleration.

### AI4INDUSTRY

An other example show how to extract GPS information per image, allowing to export each frame with meta data in filename.

* Compile
```bash
cd cmd/ai4industry
go build
```

* To see all frames info for default Ancenis video
```bash
./ai4industry
```

* To export 20 frames from 7366 to 7385 from Ancenis video:
```bash
./ai4industry -start 7366 -stop 7385 -export
```
