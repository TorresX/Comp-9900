// uuid for image name
// tutor run database 怎么办
// require('dotenv').config();
const express = require('express');
const app = express();
const bodyParser = require('body-parser');
const ejs = require('ejs');
const mongoose = require("mongoose");
const _ = require('lodash');
const session = require('express-session');
const passport = require('passport');
const passportLocalMongoose = require('passport-local-mongoose');
// dont need to require passport-local package
const GoogleStrategy = require('passport-google-oauth').OAuth2Strategy;
const findOrCreate = require('mongoose-findorcreate');
const LocalStrategy = require('passport-local').Strategy;
// image processing
//git clone  git://github.com/felixge/node-formidable.git node_modules/formidable
const formidable = require("formidable");
const fs = require("fs");

// real time post rendering
const http = require("http").createServer(app);
const io = require("socket.io")(http);

const ObjectId = require('mongodb').ObjectId;

// MD5 ashing
const md5 = require('md5');

app.use("/static", express.static(__dirname + "/static"));
app.use("/images", express.static("/Users/yingzhengwang/Pictures/propertyImages"));
app.set('view engine', 'ejs');

app.use(bodyParser.urlencoded({extended: true}));

app.use(session({
    secret: 'Rookies',// hash of a cookie sign with this secret with cookie
    resave: false,//determine whether our session would be updated even though a user may not make a change
    saveUninitialized: false, // true: create a cookie whenever a user visit even without logging in
}));

app.use(passport.initialize());
app.use(passport.session());

// connect mongo database
mongoose.connect("mongodb://localhost:27017/rookiesDB",{ useUnifiedTopology: true,useNewUrlParser: true});
mongoose.set("useCreateIndex",true);
// create a property schema
const PropertySchema = new mongoose.Schema({
   suburb: String,
    address: String,
    price: Number,
    bedroom: Number,
    bathroom: Number,
    garage: Number,
    un_available_dates: [Date],
    // owner: {type: mongoose.Schema.Types.ObjectId,
    //         ref: 'Landlord'},
    owner: String,
    images: [String],
    comments: [{username: String, comment: String}]
});

const UserSchema = new mongoose.Schema({
    _id: String,
    name: String,
    username: String,
    password: String,
    phone: String,
    googleId: String,
    token: String,
    orders: [{type: mongoose.Schema.Types.ObjectId,
        ref: 'Order'}],
});

const OrderSchema = new mongoose.Schema({
    order_dates: [Date],
    order_property: {type: mongoose.Schema.Types.ObjectId,
                        ref: 'Property'},
    order_user: {type: mongoose.Schema.Types.ObjectId,
                ref: 'User'},
});

const LandlordSchema = new mongoose.Schema({
    _id: String,
    name: String,
    username: String,
    password: String,
    phone: String,
    googleId: String,
    token: String,
    property: [{type: mongoose.Schema.Types.ObjectId,
        ref: 'Property'}]
});

UserSchema.plugin(passportLocalMongoose);
UserSchema.plugin(findOrCreate);
LandlordSchema.plugin(passportLocalMongoose);
LandlordSchema.plugin(findOrCreate);

const User = mongoose.model("user", UserSchema);
const Landlord = mongoose.model("landlord", LandlordSchema);
const Property = mongoose.model("property", PropertySchema);
const Order = mongoose.model("order",OrderSchema);

// passport.use(User.createStrategy());
// passport.use(Landlord.createStrategy());

// passport.serializeUser(function(user, done) {
//     done(null, user.id);
// });
//
// passport.deserializeUser(function(id, done) {
//     User.findById(id, function(err, user) {
//         done(err, user);
//     });
// });
//
// passport.deserializeUser(function(id, done) {
//     Landlord.findById(id, function(err, user) {
//         done(err, user);
//     });
// });

// const user = new User({
//     name: "John",
//     password: "123456",
//     email: "1@2.com",
//     phone: "0412341234",
// });
//
// user.save();

// const landlord = new Landlord({
//     name: "Simon So",
//     password: "simon123",
//     email: "SimonSo@gmail.com",
//     phone: "0448 006 242",
// });
//
// landlord.save();

// Landlord.updateOne({ name: 'Simon So' }, { googleId: "abc" }, function (err,result) {
//     if (!err){
//         console.log(result);
//     }
// });

// const property = new Property({
//     suburb: "Parramatta",
//     address: "20/33-35 Cowper Street, Parramatta, NSW 2150",
//     price: 128,
//     bedroom: 2,
//     bathroom: 2,
//     garage: 0,
//     un_available_dates:[new Date('2010-7-13')],
//     comments: [{username: "Bob", comment: "Nice property"}]
// });
//
//
// property.save();

// newimages = [
//     "static/propertyImages/bondi41.jpg",
//     "static/propertyImages/bondi42.jpg",
//     "static/propertyImages/bondi43.jpg",
//     "static/propertyImages/bondi44.jpg",
//     "static/propertyImages/bondi45.jpg"
// ];
//
// Property.updateOne({_id: ObjectId("5dc125d45307f217f2e9fd87")}, {images: newimages}, function (err,result) {
//    if (!err){
//        console.log(result);
//    }
// });

// append the owner
// Landlord.findOne({name: 'Simon So'}, function (err,ele) {
//     if (!err){
//         console.log(ele);
//         Property.updateOne({suburb: "Parramatta"},{owner: ele}, function (err, result) {
//             console.log(result)
//         })
//     }
// });

// sign up
// Property.findOne({address: "66/7-19 James st, Lidcombe, NSW 2141"},function (err,ele) {
//     ele.images.push("/static/propertyImages/2019_calender.jpeg");
//     ele.save();
// });

passport.use("userStrategy",new LocalStrategy(function (name, password,done) {
    console.log(name,password,done);
    User.findOne({ username: name }, function(err, user) {
        console.log(user);
        if (err) {
            // console.log("bbbb");
            return done(err);
        }
        if (!user) {
            // console.log("cccc");
            return done(null, false, { message: 'Incorrect username.' });
        }
        // if (!user.authenticate(password)) {
        //     // console.log("dddd");
        //     return done(null, false, { message: 'Incorrect password.' });
        // }
        if (user.password != md5(password)) {
            console.log("dddd");
            return done(null, false, { message: 'Incorrect password.' });
        }
        // console.log("eeee");
        return done(null, user);
    });
}));

// define Landlord Strategy
passport.use("landlordStrategy",new LocalStrategy(function (name, password,done) {
    Landlord.findOne({ username: name }, function(err, landlord) {
        console.log(`lanlordStrategy: ${landlord}`);
        if (err) {
            // console.log("bbbb");
            return done(err);
        }
        if (!landlord) {
            // console.log("cccc");
            return done(null, false, { message: 'Incorrect username.' });
        }
        if (landlord.password != md5(password)) {
            // console.log("dddd");
            return done(null, false, { message: 'Incorrect password.' });
        }
        // console.log("eeee");
        // console.log(`password: ${password}`);

        // if (!landlord.validPassword(password)){
        //     console.log( 'Incorrect password.' );
        // }
        return done(null, landlord);
    });
}));

passport.serializeUser(function(user, done) {
    // console.log("serializeUser");
    // console.log(user.id);
    done(null, user.id);
});

passport.deserializeUser(function(id, done) {
    console.log("DeserializeUser");
    console.log(id);
    if (_.startsWith(id, 'user')){
        User.findById(id, function(err, user) {
            // console.log(err);
            // console.log(Object.getPrototypeOf(user). collection.name);
            console.log(user);
            done(err, user);
        });
    }else{
        Landlord.findById(id, function(err, user) {
            // console.log(err);
            // console.log(Object.getPrototypeOf(user). collection.name);
            console.log(user);
            done(err, user);
        });
    }

});

http.listen(3000, function (){
    console.log("Server started at 3000");
});

app.route("/")
    .get( function (req, res) {
        res.render("home.ejs");
    });

app.route("/signup")
    .get(function (req, res) {
        if (!req.user){
            // a user has not logged in yet, render the login page
            res.render("signup.ejs");
        }else{
            if (_.startsWith(req.user.id,"user")){
                res.redirect("/user");
            }else{
                res.redirect("landlord");
            }
        }
    })
    .post(function (req, res) {
        // console.log(req.body);
        // DOM already checked if the two passwords are matched
        // if (req.body.password != req.body.re_password){
        //     res.redirect("/signup");
        // }
        // check if it is landlord or not
        if (req.body.iamlandlord === 'on'){ // landlord
            // check if it is already in database
            Landlord.findOne({email: req.body.username}, function (err, returnedPerson) {
                if (!err){
                    if (!returnedPerson){
                        // no error and no existing landlord
                        // register
                        const newUser = new Landlord({
                            _id: "landlord"+ req.body.username,
                            name: req.body.name,
                            username: req.body.username,
                            password: md5(req.body.password)
                        });
                        // Landlord.register(newUser, req.body.password, function (err,user) {
                        //     if (err){
                        //         console.log(err);
                        //         res.redirect("/signup");
                        //     }else{
                        //         passport.authenticate("landlordStrategy")(req, res, function () {
                        //             res.redirect("/landlord");
                        //         });
                        //     }
                        // })

                        // 7/11/2019 update signin function
                        newUser.save(function (err, user, numAffected) {
                            if (user){
                                passport.authenticate("landlordStrategy")(req, res, function () {
                                    res.redirect("/landlord");
                                });
                            }
                        });

                    }else{
                        // found a already existed landlord
                    }
                }
            })
        }else{
            User.findOne({email: req.body.username}, function (err, returnedPerson) {
                if (!err){
                    if (!returnedPerson){
                        // no error and no existing user
                        // register
                        const newUser = new User({
                            _id: "user"+ req.body.username,
                            username: req.body.username,
                            name: req.body.name,
                            password: md5(req.body.password)
                        });
                        // console.log(newUser);
                        // User.register(newUser, req.body.password, function (err,user) {
                        //     if (err){
                        //         console.log(err);
                        //         res.redirect("/signup");
                        //     }else{
                        //         // console.log("aaaaaaaaaaaaa");
                        //         passport.authenticate("userStrategy")(req, res, function () {
                        //             // console.log("bbbbbbbbbb");
                        //             res.redirect("/user");
                        //         });
                        //     }
                        // })

                        // update login
                        User.insertMany([newUser], function(error, docs) {
                            passport.authenticate("userStrategy")(req, res, function () {
                                // console.log("bbbbbbbbbb");
                                res.redirect("/user");
                        });
                        });
                    }else{
                        // found a already existed user
                    }
                }
            })
        }
        // check the user's email has already been signed up

        //
    });

app.route("/login")
    .get(function (req, res) {
        if (!req.user){
            // a user has not logged in yet, render the login page
            res.render("login.ejs");
        }else{
            if (_.startsWith(req.user.id,"user")){
                res.redirect("/user");
            }else{
                res.redirect("landlord");
            }
        }
    })
    .post(function (req, res) {
        const lookingForUsername = req.body.username;
        if (req.body.iamlandlord === 'on'){
            // check landlord collection
            const user = new Landlord({
                _id: "landlord"+ lookingForUsername,
                username: lookingForUsername,
                password: req.body.password
            });
            console.log(`landlord: ${user}`);
            req.login(user, function (err) {
                if (err){
                    return next(err);
                }else{
                    console.log("landlord loginauthenticate");
                    passport.authenticate("landlordStrategy")(req, res, function () {
                        res.redirect("/landlord");
                    })
                }
            });
        }else{
            // check user collection
            const user = new User({
                _id: "user"+ lookingForUsername,
                username: lookingForUsername,
                password: req.body.password
            });
            req.login(user, function (err) {
                if (err){
                    return next(err);
                }else{
                    // console.log("user loginauthenticate");
                    passport.authenticate("userStrategy")(req, res, function () {
                        res.redirect("/user");
                    })
                }
            })
        }
    });

app.route("/user")
    .get(function (req, res) {
        if (req.user){
            if (_.startsWith(req.user.id,"landlord")){
                res.redirect("/landlord");
            }else{
                if (req.isAuthenticated()){
                    // console.log("user authenticated");
                    res.render("userAcc/userAcc.ejs")
                } else{
                    res.redirect('/login');
                }
            }
        }else{
            res.redirect('/login');
        }
    });

app.route("/landlord")
    .get(function (req, res) {
        if (req.user){
            if (_.startsWith(req.user.id,"user")){
                res.redirect("/user");
            }else{
                // console.log(req.user);
                if (req.isAuthenticated()){
                    // console.log("landlord authenticated");
                    res.render("landlordAcc/dashboard.ejs",{landlordName: req.user.name});
                } else{
                    res.redirect('/login');
                }
            }
        }else {
            res.redirect('/login');
        }
    });

// landlordAcc icon in the header
app.route("/landlordAcc")
    .get(function (req, res) {
        if (!req.user){
            // a user has not logged in yet
            res.redirect("/login");
        }else{
            if (_.startsWith(req.user.id,"user")){
                res.redirect("/user");
            }else{
                res.redirect("landlord");
            }
        }
    });

// logout anckor tag in the *Acc.ejs file
app.get("/logout", function (req, res) {
    req.logout();
    res.redirect('/');
});

app.post("/properties", function (req, res) {
    Property.find({}, function (err,properties) {
        console.log(properties);
        res.render("property/properties.ejs",{properties: properties});
    });

});

app.get("/properties", function (req, res) {
    Property.find({}, function (err,properties) {
        console.log(properties);
        res.render("property/properties.ejs",{properties: properties});
    });

});

app.get("/properties/:propertyId", function (req, res) {
    console.log(req.params.propertyId);
    Property.findOne({_id: req.params.propertyId}, function (err,property) {
        console.log(property);
        res.render("property/property.ejs",{property: property});
    });

});

app.get("/myproperties", function (req, res) {

    if ((req.user) && (_.startsWith(req.user.id, "landlord"))){
        console.log(req.user);
        Property.find({owner: {$eq: req.user.id}}, function (err,properties) {
            res.render("landlordAcc/myproperties",{properties: properties});
        });
    }else{
        res.redirect("/");
    }
});

io.on("connection", function (socket) {
    console.log("User connected");
    socket.on("new_post", function (formData) { // listen to the event
        // console.log(formData);
        socket.broadcast.emit("new_post", formData);
    });
});

app.get("/posts", function (req, res) {
    if ((req.user) && (_.startsWith(req.user.id, "landlord"))){
        res.render("landlordAcc/posts.ejs")
    }else{
        res.redirect("/");
    }
});

app.post("/do-post", function (req, res) {
    // { address: '18/12 Evans Avenue',
    //     suburb: 'Parramatta',
    //     bedroom: '1',
    //     bathroom: '1',
    //     garage: '1',
    //     price: '99',
    //     image: '/static/propertyImages/5672a5e1a3299.jpg' }
    // { property: [],
    //     _id: 'landlordaaa@gmail.com',
    //     name: 'Yingzheng Wang',
    //     username: 'aaa@gmail.com',
    //     __v: 0 }
    // owner: {type: mongoose.Schema.Types.ObjectId,
    //     ref: 'Landlord'},

    console.log(req.user);
    console.log(req.body);
    if (req.user){
        postProperty = new Property({
            address: req.body.address,
            suburb: req.body.suburb,
            bedroom: req.body.bedroom,
            bathroom: req.body.bathroom,
            garage: req.body.garage,
            price: req.body.price,
            owner: req.user._id,
            images: req.body.images
        });
        // postProperty.save();
        Property.insertMany([postProperty], function(error, docs) {
            res.send({
                text: "posted successfully",
                _id: docs[0].id
            })
        });

        // 在存到mongodb里面去的时候，由于存放过程太慢，读取过程很快，所以程序执行到这里的时候
        // 上面的数据还没存进mongodb， 那么读出来的数据必定为空
        // Property.findOne({address: req.body.address}, function (err,property) {
        //     console.log(property);
        //     // res.send({
        //     //     text: "posted successfully",
        //     //     _id: property._id,
        //     //
        //     // });
        // });

    }
});

app.post("/do-upload-image", function (req, res) {
    // create a new instance of form data
    console.log(req.files);
    const formData = new formidable.IncomingForm();
    var imagesPath = [];
    formData.parse(req, function (error, fields, files) {
        // here files is the name of filed set in bootstrap model
        // old path is the path which user selected
        // and the new path where selected image will be stored
      if (files.file1.size != 0){
          var oldPath = files.file1.path;
          var newPath = "static/propertyImages/"+ files.file1.name;
          imagesPath.push(newPath);
          fs.rename(oldPath,newPath, function (err) {
              // successfully store the image
              // stored in data base via ajax in posts.ejs
              // console.log(newPath);
              if (err){
                  console.log("can not save file 1")
              }
          });
      }
        if (files.file2.size != 0){
            var oldPath = files.file2.path;
            var newPath = "static/propertyImages/"+ files.file2.name;
            imagesPath.push(newPath);
            fs.rename(oldPath,newPath, function (err) {
                // successfully store the image
                // stored in data base via ajax in posts.ejs
                // console.log(newPath);
                if (err){
                    console.log("can not save file 2")
                }
            });
        }

        if (files.file3.size != 0){
            var oldPath = files.file3.path;
            var newPath = "static/propertyImages/"+ files.file3.name;
            imagesPath.push(newPath);
            fs.rename(oldPath,newPath, function (err) {
                // successfully store the image
                // stored in data base via ajax in posts.ejs
                // console.log(newPath);
                if (err){
                    console.log("can not save file 3")
                }
            });
        }

        if (files.file4.size != 0){
            var oldPath = files.file4.path;
            var newPath = "static/propertyImages/"+ files.file4.name;
            imagesPath.push(newPath);
            fs.rename(oldPath,newPath, function (err) {
                // successfully store the image
                // stored in data base via ajax in posts.ejs
                // console.log(newPath);
                if (err){
                    console.log("can not save file 4")
                }
            });
        }
        if (files.file5.size != 0){
            var oldPath = files.file5.path;
            var newPath = "static/propertyImages/"+ files.file5.name;
            imagesPath.push(newPath);
            fs.rename(oldPath,newPath, function (err) {
                // successfully store the image
                // stored in data base via ajax in posts.ejs
                // console.log(newPath);
                if (err){
                    console.log("can not save file 5")
                }
            });
        }

        if (files.file6.size != 0){
            var oldPath = files.file6.path;
            var newPath = "static/propertyImages/"+ files.file6.name;
            imagesPath.push(newPath);
            fs.rename(oldPath,newPath, function (err) {
                // successfully store the image
                // stored in data base via ajax in posts.ejs
                // console.log(newPath);
                if (err){
                    console.log("can not save file 6")
                }
            });
        }

        if (files.file7.size != 0){
            var oldPath = files.file7.path;
            var newPath = "static/propertyImages/"+ files.file7.name;
            imagesPath.push(newPath);
            fs.rename(oldPath,newPath, function (err) {
                // successfully store the image
                // stored in data base via ajax in posts.ejs
                // console.log(newPath);
                if (err){
                    console.log("can not save file 7")
                }
            });
        }

        if (files.file8.size != 0){
            var oldPath = files.file8.path;
            var newPath = "static/propertyImages/"+ files.file8.name;
            imagesPath.push(newPath);
            fs.rename(oldPath,newPath, function (err) {
                // successfully store the image
                // stored in data base via ajax in posts.ejs
                // console.log(newPath);
                if (err){
                    console.log("can not save file 8")
                }
            });
        }
        // console.log(imagesPath);
        res.send(imagesPath);
        // const oldPath = files.file.path;
        // const newPath = "static/propertyImages/"+ files.file.name;
        // // will upload the file in that folder
        // fs.rename(oldPath,newPath, function (err) {
        //     // successfully store the image
        //     // stored in data base via ajas in posts.ejs
        //     console.log("fs send back response")
        //     res.send("/"+ newPath);
        // });
    });


});

